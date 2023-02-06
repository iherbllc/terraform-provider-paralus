// Utility methods for testing
package utils

import (
	"fmt"

	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

// Validate that the passed in map of project roles match what was set in the project struct
func ValidateProjectNamespaceRolesSet(pnrStruct []*userv3.ProjectNamespaceRole, projectRoles map[string]string) error {
	if pnrStruct == nil {
		return fmt.Errorf("project namespace roles list does not exist")
	}
	if len(pnrStruct) <= 0 {
		return fmt.Errorf("project namespace roles list is empty")
	}

	checkMap := make(map[string]string)

	// loop through roles
	for _, role := range pnrStruct {
		// loop through passed in user roles
		for key, val := range projectRoles {
			// set the value once if it doesn't exist
			if _, exists := checkMap[key]; !exists {
				checkMap[key] = val
			}
			switch key {
			case "role":
				if role.Role == projectRoles["role"] {
					checkMap["role"] = ""
				}
			case "project":
				if *role.Project == projectRoles["project"] {
					checkMap["project"] = ""
				}
			case "namespace":
				if *role.Namespace == projectRoles["namespace"] {
					checkMap["namespace"] = ""
				}
			case "group":
				if *role.Group == projectRoles["group"] {
					checkMap["group"] = ""
				}
			}
		}
	}

	return buildResultError(checkMap, projectRoles)
}

// Validate that the passed in map of user roles match what was set in the project struct
func ValidateUserRolesSet(urStruct []*userv3.UserRole, userRoles map[string]string) error {
	if urStruct == nil {
		return fmt.Errorf("user roles list does not exist")
	}
	if len(urStruct) <= 0 {
		return fmt.Errorf("user roles list is empty")
	}

	checkMap := make(map[string]string)

	// loop through roles
	for _, role := range urStruct {
		// loop through passed in user roles
		for key, val := range userRoles {
			// set the value once if it doesn't exist
			if _, exists := checkMap[key]; !exists {
				checkMap[key] = val
			}
			switch key {
			case "role":
				if role.Role == userRoles["role"] {
					checkMap["role"] = ""
				}
			case "user":
				if role.User == userRoles["user"] {
					checkMap["user"] = ""
				}
			case "namespace":
				if role.Namespace == userRoles["namespace"] {
					checkMap["namespace"] = ""
				}
			}
		}
	}
	return buildResultError(checkMap, userRoles)
}

// Build result error message from maps
func buildResultError(filledMap map[string]string, sourceMap map[string]string) error {
	var invalidStr string
	var invalidVlStr string
	index := 1
	for key, val := range filledMap {
		if val != "" {
			if index > 1 {
				invalidStr += ","
				invalidVlStr += ","
			}
			invalidStr += key
			invalidVlStr += sourceMap[key]
			index++
		}
	}
	if invalidStr != "" {
		return fmt.Errorf("invalid %s: %s", invalidStr, invalidVlStr)
	}
	return nil
}
