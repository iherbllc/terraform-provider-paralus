// Utility methods for PCTL Group struct
package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

// Build the group struct from a schema resource
func BuildGroupStructFromResource(d *schema.ResourceData) *userv3.Group {

	groupStruct := userv3.Group{
		Kind: "Group",
		Metadata: &commonv3.Metadata{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		},
		Spec: &userv3.GroupSpec{},
	}

	if projectRoles, ok := d.GetOk("project_roles"); ok {
		groupStruct.Spec.ProjectNamespaceRoles = make([]*userv3.ProjectNamespaceRole, 0)
		rolesList := projectRoles.([]interface{})
		for _, eachRole := range rolesList {
			if role, ok := eachRole.(map[string]interface{}); ok {
				namespace := role["namespace"].(string)
				group := role["group"].(string)
				project := role["project"].(string)
				groupStruct.Spec.ProjectNamespaceRoles = append(groupStruct.Spec.ProjectNamespaceRoles, &userv3.ProjectNamespaceRole{
					Project:   &project,
					Role:      role["role"].(string),
					Namespace: &namespace,
					Group:     &group,
				})
			}
		}
	}

	if users, ok := d.GetOk("users"); ok {
		usersList := users.([]interface{})
		groupStruct.Spec.Users = make([]string, len(usersList))
		for i, v := range usersList {
			groupStruct.Spec.Users[i] = v.(string)
		}
	}

	if groupType, ok := d.GetOk("type"); ok {
		groupStruct.Spec.Type = groupType.(string)
	}

	return &groupStruct
}

// Build the schema resource from group Struct
func BuildResourceFromGroupStruct(group *userv3.Group, d *schema.ResourceData) {
	d.Set("name", group.Metadata.Name)
	d.Set("description", group.Metadata.Description)
	projectRoles := make([]map[string]interface{}, 0)
	for _, role := range group.Spec.GetProjectNamespaceRoles() {
		projectRoles = append(projectRoles, map[string]interface{}{
			"project":   role.Project,
			"role":      role.Role,
			"namespace": role.Namespace,
			"group":     role.Group,
		})
	}
	d.Set("project_roles", projectRoles)
	d.Set("users", group.Spec.Users)
	d.Set("type", group.Spec.Type)
}
