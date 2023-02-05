// Testing utils package
package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/pkg/errors"
)

// AssertStringNotEmpty asserts when the string is not empty
func AssertStringNotEmpty(message, str string) diag.Diagnostics {
	var diags diag.Diagnostics
	str = strings.TrimSpace(str)
	if str != "" {
		return diags
	}

	if message != "" {
		return diag.FromErr(errors.New(fmt.Sprintf("%s: expected not empty string.", message)))
	} else {
		return diag.FromErr(errors.New("Expected not empty string."))
	}
}

func ValidateProjectNamespaceRolesSet(pnrStruct []*userv3.ProjectNamespaceRole, projectRoles map[string]string) error {
	if pnrStruct == nil {
		return fmt.Errorf("project Namespace Roles list does not exist")
	}
	if len(pnrStruct) <= 0 {
		return fmt.Errorf("project Namespace Roles list is empty")
	}

	mapKey := reflect.ValueOf(projectRoles).MapKeys()

	for _, role := range pnrStruct {
		switch key := mapKey[0].Interface().(string); key {
		case "role":
			if role.Role != projectRoles["role"] {
				return fmt.Errorf("invalid role: %s", projectRoles["role"])
			}
		case "project":
			if *role.Project != projectRoles["project"] {
				return fmt.Errorf("invalid project: %s", projectRoles["project"])
			}
		case "namespace":
			if *role.Project != projectRoles["namespace"] {
				return fmt.Errorf("invalid namespace: %s", projectRoles["namespace"])
			}
		case "group":
			if *role.Project != projectRoles["group"] {
				return fmt.Errorf("invalid group: %s", projectRoles["group"])
			}
		}
	}
	return nil
}

// Check to make sure that the list of roles from ProjectNamespaceRole is unique.
// This is required due to a limitation with Paralus.
// See: https://github.com/paralus/paralus/issues/136
func AssertUniqueRoles(pnrStruct []*userv3.ProjectNamespaceRole) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(pnrStruct) >= 2 {
		pnrStructMap := make(map[string]string)
		for _, role := range pnrStruct {
			roleInfo := fmt.Sprintf("%s, %s, %s, %s", *role.Group, *role.Namespace, role.Role, *role.Project)
			if _, exists := pnrStructMap[roleInfo]; exists {
				return diag.FromErr(errors.New("permissions must be distinct between project_roles blocks. Add the project permissions to the group directly if duplicates are required."))
			}
			pnrStructMap[roleInfo] = "unique"
		}

	}

	return diags
}
