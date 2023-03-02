// Utility methods for PCTL User struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/paralus/cli/pkg/authprofile"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/pkg/errors"
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
