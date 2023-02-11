// Utility methods for PCTL User struct
package utils

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/paralus/cli/pkg/user"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

// Check users from a list exist in paralus
func CheckUsersExist(users []string) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(users) > 0 {
		for _, usr := range users {
			userStruct, _ := GetUserByName(usr)
			if userStruct == nil {
				return diag.FromErr(fmt.Errorf("user '%s' does not exist", usr))
			}
		}
	}
	return diags
}

// Check users from a UserRole structs exist in paralus
func CheckUserRoleUsersExist(userRoles []*userv3.UserRole) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(userRoles) > 0 {
		for _, userRole := range userRoles {
			userStruct, _ := user.GetUserByName(userRole.User)
			if userStruct == nil {
				return diag.FromErr(fmt.Errorf("user '%s' does not exist", userRole.User))
			}
		}
	}
	return diags
}

// Get user by name
func GetUserByName(userName string) (*userv3.User, error) {
	uri := fmt.Sprintf("/auth/v3/user/%s", userName)
	resp, err := makeRestCall(uri, "GET", nil)
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
