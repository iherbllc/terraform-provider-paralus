// Utility methods for PCTL Project struct
package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

// Build the project struct from a schema resource
func BuildProjectStructFromResource(d *schema.ResourceData) *systemv3.Project {

	projectStruct := systemv3.Project{
		Kind: "Project",
		Metadata: &commonv3.Metadata{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Id:          d.Get("uuid").(string),
		},
		Spec: &systemv3.ProjectSpec{},
	}
	// define project roles
	if projectRoles, ok := d.GetOk("project_roles"); ok {
		projectStruct.Spec.ProjectNamespaceRoles = make([]*userv3.ProjectNamespaceRole, 0)
		rolesList := projectRoles.([]interface{})
		for _, eachRole := range rolesList {
			if role, ok := eachRole.(map[string]interface{}); ok {
				namespace := role["namespace"].(string)
				group := role["group"].(string)
				projectStruct.Spec.ProjectNamespaceRoles = append(projectStruct.Spec.ProjectNamespaceRoles, &userv3.ProjectNamespaceRole{
					Project:   &projectStruct.Metadata.Name,
					Role:      role["role"].(string),
					Namespace: &namespace,
					Group:     &group,
				})
			}
		}
	}
	// define user roles
	if userRoles, ok := d.GetOk("user_roles"); ok {
		projectStruct.Spec.UserRoles = make([]*userv3.UserRole, 0)
		rolesList := userRoles.([]interface{})
		for _, eachRole := range rolesList {
			if role, ok := eachRole.(map[string]interface{}); ok {
				projectStruct.Spec.UserRoles = append(projectStruct.Spec.UserRoles, &userv3.UserRole{
					User: role["user"].(string),
					Role: role["role"].(string),
				})
			}
		}
	}

	return &projectStruct
}

// Build the schema resource from project Struct
func BuildResourceFromProjectStruct(project *systemv3.Project, d *schema.ResourceData) {
	d.Set("name", project.Metadata.Name)
	d.Set("description", project.Metadata.Description)
	d.Set("uuid", project.Metadata.Id)
	projectRoles := make([]map[string]interface{}, 0)
	for _, role := range project.Spec.GetProjectNamespaceRoles() {
		projectRoles = append(projectRoles, map[string]interface{}{
			"project":   role.Project,
			"role":      role.Role,
			"namespace": role.Namespace,
			"group":     role.Group,
		})
	}
	d.Set("project_roles", projectRoles)
	userRoles := make([]map[string]interface{}, 0)
	for _, role := range project.Spec.GetUserRoles() {
		userRoles = append(userRoles, map[string]interface{}{
			"user": role.User,
			"role": role.Role,
		})
	}
	d.Set("user_roles", userRoles)
}
