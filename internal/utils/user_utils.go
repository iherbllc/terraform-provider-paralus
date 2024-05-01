// Utility methods for PCTL User struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
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
					diags.AddError(fmt.Sprintf("user '%s' does not exist", usr), "")
					return diags
				}
				diags.AddError(fmt.Sprintf("error getting user %s info", usr), err.Error())
				return diags
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
					diags.AddError(fmt.Sprintf("user '%s' does not exist", userRole.User), "")
					return diags
				}

				diags.AddError(fmt.Sprintf("error getting user %s info", userRole.User), err.Error())
				return diags
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

func BuildKubeConfigStruct(ctx context.Context, data *structs.KubeConfig, kubeconfigYAML string) (string, diag.Diagnostics) {
	// decode := k8Scheme.Codecs.UniversalDeserializer().Decode
	// obj, _, err := decode([]byte(kubeconfigYAML), nil, nil)
	var diagsReturn diag.Diagnostics
	var diags diag.Diagnostics
	config, err := clientcmd.Load([]byte(kubeconfigYAML))

	if err != nil {
		diagsReturn.AddError("Error loading kubeconfigYAML", err.Error())
		return kubeconfigYAML, diagsReturn
	}

	if config.AuthInfos != nil {
		for _, authInfo := range config.AuthInfos {
			data.ClientCertificateData = types.StringValue(string(authInfo.ClientCertificateData))
			data.ClientKeyData = types.StringValue(string(authInfo.ClientKeyData))
			break
		}
	}

	if config.Clusters != nil {
		clusters := make([]structs.ClusterInfo, 0)
		for _, clusterInfo := range config.Clusters {
			clusters = append(clusters, structs.ClusterInfo{
				CertificateAuthorityData: types.StringValue(clusterInfo.CertificateAuthority),
				Server:                   types.StringValue(clusterInfo.Server),
			})
		}

		data.ClusterInfo, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: structs.ClusterInfo{}.AttributeTypes()}, clusters)
		diagsReturn.Append(diags...)
	}

	return "", nil
}

// Build the project struct from a schema resource
func BuildUsersStructFromResource(ctx context.Context, data *structs.User) ([]*userv3.User, diag.Diagnostics) {

	users := make([]*userv3.User, 0)
	if !data.UsersInfo.IsNull() {
		usersInfo := make([]structs.UserInfo, 0, len(data.UsersInfo.Elements()))
		diags := data.UsersInfo.ElementsAs(ctx, &usersInfo, false)
		if diags.HasError() {
			return nil, diags
		}
		for _, user := range usersInfo {
			userStruct := &userv3.User{
				Kind: "User",
				Metadata: &commonv3.Metadata{
					Name: user.Email.ValueString(),
					Id:   user.Id.ValueString(),
				},
				Spec: &userv3.UserSpec{
					FirstName: user.FirstName.ValueString(),
					LastName:  user.LastName.ValueString(),
				},
			}
			if !user.Groups.IsNull() {
				groups := make([]types.String, len(user.Groups.Elements()))
				diags := user.Groups.ElementsAs(ctx, &groups, false)
				if diags.HasError() {
					return nil, diags
				}
				for _, v := range groups {
					userStruct.Spec.Groups = append(userStruct.Spec.Groups, v.ValueString())
				}
			}

			// define project roles
			if !user.ProjectRoles.IsNull() {

				projectRoles := make([]structs.ProjectRole, 0, len(user.ProjectRoles.Elements()))
				diags := user.ProjectRoles.ElementsAs(ctx, &projectRoles, false)
				if diags.HasError() {
					return nil, diags
				}
				userStruct.Spec.ProjectNamespaceRoles = make([]*userv3.ProjectNamespaceRole, 0)
				for _, projectRole := range projectRoles {
					project := projectRole.Project.ValueString()
					namespace := projectRole.Namespace.ValueString()
					group := projectRole.Group.ValueString()
					userStruct.Spec.ProjectNamespaceRoles = append(userStruct.Spec.ProjectNamespaceRoles, &userv3.ProjectNamespaceRole{
						Project:   &project,
						Role:      projectRole.Role.ValueString(),
						Namespace: &namespace,
						Group:     &group,
					})
				}
			}
			users = append(users, userStruct)
		}
	}
	return users, nil
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
func BuildResourceFromUsersStruct(ctx context.Context, users []*userv3.User, data *structs.User) diag.Diagnostics {
	var diagsReturn diag.Diagnostics
	var diags diag.Diagnostics

	if len(users) == 0 {
		return diagsReturn
	}

	usersInfo := make([]structs.UserInfo, 0)
	for _, user := range users {
		projectRoles := make([]structs.ProjectRole, 0)
		for _, role := range user.Spec.GetProjectNamespaceRoles() {
			projectRoles = append(projectRoles, structs.ProjectRole{
				Project:   types.StringValue(DerefString(role.Project)),
				Role:      types.StringValue(role.Role),
				Namespace: types.StringValue(DerefString(role.Namespace)),
				Group:     types.StringValue(DerefString(role.Group)),
			})
		}
		userInfo := structs.UserInfo{
			FirstName: types.StringValue(user.Spec.FirstName),
			LastName:  types.StringValue(user.Spec.LastName),
			Email:     types.StringValue(user.Metadata.Name),
			Id:        types.StringValue(user.Metadata.Id),
		}
		userInfo.ProjectRoles, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: structs.ProjectRole{}.AttributeTypes()}, projectRoles)
		diagsReturn.Append(diags...)
		userInfo.Groups, diags = types.ListValueFrom(ctx, types.StringType, user.Spec.Groups)
		diagsReturn.Append(diags...)

		usersInfo = append(usersInfo, userInfo)
	}

	data.UsersInfo, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: structs.UserInfo{}.AttributeTypes()}, usersInfo)
	diagsReturn.Append(diags...)

	return diagsReturn
}
