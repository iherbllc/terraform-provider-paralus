// Utility methods for PCTL User struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
)

// Check users from a list exist in paralus
func CheckUsersExist(ctx context.Context, users []string, auth *authprofile.Profile) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(users) > 0 {
		for _, usr := range users {
			_, err := GetUserByName(ctx, usr, auth)
			if err != nil {
				if err == ErrResourceNotExists {
					return diag.FromErr(fmt.Errorf("user '%s' does not exist", usr))
				}
				return diag.FromErr(errors.Wrapf(err, "error getting user %s info", usr))
			}
		}
	}
	return diags
}

// Check users from a UserRole structs exist in paralus
func CheckUserRoleUsersExist(ctx context.Context, userRoles []*userv3.UserRole, auth *authprofile.Profile) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(userRoles) > 0 {
		for _, userRole := range userRoles {
			_, err := GetUserByName(ctx, userRole.User, auth)
			if err != nil {
				if err == ErrResourceNotExists {
					return diag.FromErr(fmt.Errorf("user '%s' does not exist", userRole.User))
				}
				return diag.FromErr(errors.Wrapf(err, "error getting user %s info", userRole.User))
			}
		}
	}
	return diags
}

// Get all users based on provided filter/limit/offset
func GetUsers(ctx context.Context, params []string, auth *authprofile.Profile) ([]*userv3.User, error) {

	uri := "/auth/v3/users?" + strings.TrimSuffix(strings.Join(params, "&"), "&")
	resp, err := makeRestCall(ctx, uri, "GET", nil, auth)
	if err != nil {
		return nil, err
	}
	userList := &userv3.UserList{}
	err = json.Unmarshal([]byte(resp), userList)
	if err != nil {
		return nil, err
	}

	return userList.Items, nil

}

// Get user by name
func GetUserByName(ctx context.Context, userName string, auth *authprofile.Profile) (*userv3.User, error) {
	uri := fmt.Sprintf("/auth/v3/user/%s", userName)
	resp, err := makeRestCall(ctx, uri, "GET", nil, auth)
	if err != nil {
		return nil, err
	}
	user := &userv3.User{}
	err = json.Unmarshal([]byte(resp), user)
	if err != nil {
		return nil, err
	}

	return user, nil

}

// retrieves the kubeconfig for the user with either all or specific cluster info
func GetKubeConfig(ctx context.Context, accountID string, namespace string, cluster string, auth *authprofile.Profile) (string, error) {
	params := url.Values{}
	if namespace != "" {
		params.Add("namespace", namespace)
	}
	if cluster != "" {
		params.Add("opts.selector", fmt.Sprintf("paralus.dev/clusterName=%s", cluster))
	}
	params.Add("opts.account", accountID)
	params.Add("opts.organization", config.GetConfig().Organization)
	uri := fmt.Sprintf("/v2/sentry/kubeconfig/user?%s", params.Encode())
	return makeRestCall(ctx, uri, "GET", nil, auth)
}

func BuildKubeConfigStruct(ctx context.Context, d *schema.ResourceData, kubeconfigYAML string) (string, error) {
	// decode := k8Scheme.Codecs.UniversalDeserializer().Decode
	// obj, _, err := decode([]byte(kubeconfigYAML), nil, nil)
	config, err := clientcmd.Load([]byte(kubeconfigYAML))

	if err != nil {
		return kubeconfigYAML, err
	}

	if config.AuthInfos != nil {
		for _, authInfo := range config.AuthInfos {
			d.Set("client_certificate_data", string(authInfo.ClientCertificateData))
			d.Set("client_key_data", string(authInfo.ClientKeyData))
			break
		}
	}

	if config.Clusters != nil {
		clusters := make([]map[string]interface{}, 0)
		for _, clusterInfo := range config.Clusters {
			clusters = append(clusters, map[string]interface{}{
				"certificate_authority_data": clusterInfo.CertificateAuthority,
				"server":                     clusterInfo.Server,
			})
		}
		d.Set("cluster_info", clusters)
	}

	return "", nil
}

// Build the project struct from a schema resource
func BuildUsersStructFromResource(d *schema.ResourceData) []*userv3.User {

	users := make([]*userv3.User, 0)
	if usersInfo, ok := d.GetOk("users_info"); ok {
		usersList := usersInfo.([]interface{})
		for _, eachUser := range usersList {
			if user, ok := eachUser.(map[string]interface{}); ok {
				userStruct := &userv3.User{
					Kind: "User",
					Metadata: &commonv3.Metadata{
						Name: user["email"].(string),
						Id:   user["id"].(string),
					},
					Spec: &userv3.UserSpec{
						FirstName: user["first_name"].(string),
						LastName:  user["last_name"].(string),
						Groups:    user["groups"].([]string),
					},
				}
				// define project roles
				if projectRoles, ok := user["project_roles"]; ok {
					userStruct.Spec.ProjectNamespaceRoles = make([]*userv3.ProjectNamespaceRole, 0)
					rolesList := projectRoles.([]interface{})
					for _, eachRole := range rolesList {
						if role, ok := eachRole.(map[string]interface{}); ok {
							project := role["project"].(string)
							namespace := role["namespace"].(string)
							group := role["group"].(string)
							userStruct.Spec.ProjectNamespaceRoles = append(userStruct.Spec.ProjectNamespaceRoles, &userv3.ProjectNamespaceRole{
								Project:   &project,
								Role:      role["role"].(string),
								Namespace: &namespace,
								Group:     &group,
							})
						}
					}
				}
				users = append(users, userStruct)
			}
		}
	}
	return users
}

// Filter the list of users based on the filter value requested
func FilterUsers(users []*userv3.User, filter_is string, filter_is_value string,
	filter_is_case_senstive bool, filter_allow_more_than_one bool) ([]*userv3.User, error) {

	orig_filter_value := filter_is_value
	if !filter_is_case_senstive {
		filter_is_value = strings.ToLower(filter_is_value)
	}

	filtered_users := make([]*userv3.User, 0)

	for _, user := range users {
		userFound := false
		switch filter_is {
		case "email":
			if user.Metadata.Name == filter_is_value ||
				(!filter_is_case_senstive && strings.ToLower(user.Metadata.Name) == filter_is_value) {
				userFound = true
			}
		case "first_name":
			if user.Spec.FirstName == filter_is_value ||
				(!filter_is_case_senstive && strings.ToLower(user.Spec.FirstName) == filter_is_value) {
				userFound = true
			}
		case "last_name":
			if user.Spec.LastName == filter_is_value ||
				(!filter_is_case_senstive && strings.ToLower(user.Spec.LastName) == filter_is_value) {
				userFound = true
			}
		default:
			return nil, fmt.Errorf("unknown filter type: %s", filter_is)
		}
		if userFound {
			if len(filtered_users) > 1 && !filter_allow_more_than_one {
				return nil, fmt.Errorf("more than one user was found using the specified filter "+
					"'%s' with value '%s' and case_sensetive = %t", filter_is, orig_filter_value, filter_is_case_senstive)
			}
			filtered_users = append(filtered_users, user)
		}
	}

	if len(filtered_users) == 0 {
		return nil, fmt.Errorf("no user was found using the specified filter "+
			"'%s' with value '%s' and case_sensitive = %t", filter_is, orig_filter_value, filter_is_case_senstive)
	}

	return filtered_users, nil
}

// Build the schema resource from users Struct
func BuildResourceFromUsersStruct(users []*userv3.User, d *schema.ResourceData) {

	if len(users) == 0 {
		return
	}

	usersInfo := make([]map[string]interface{}, 0)
	for _, user := range users {
		projectRoles := make([]map[string]interface{}, 0)
		for _, role := range user.Spec.GetProjectNamespaceRoles() {
			projectRoles = append(projectRoles, map[string]interface{}{
				"project":   role.Project,
				"role":      role.Role,
				"namespace": role.Namespace,
				"group":     role.Group,
			})
		}
		usersInfo = append(usersInfo, map[string]interface{}{
			"first_name":    user.Spec.FirstName,
			"last_name":     user.Spec.LastName,
			"email":         user.Metadata.Name,
			"id":            user.Metadata.Id,
			"groups":        user.Spec.Groups,
			"project_roles": projectRoles,
		})
	}
	d.Set("users_info", usersInfo)
}
