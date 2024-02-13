// Utility methods for PCTL User struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
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
