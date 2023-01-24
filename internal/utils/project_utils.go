package utils

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

// Build a project struct from a resource
func BuildProjectStructFromString(projectStr string, project *systemv3.Project) error {
	// Need to take json project and convert to the new version
	projectBytes := []byte(projectStr)
	if err := json.Unmarshal(projectBytes, &project); err != nil {
		return err
	}

	return nil
}

// Build a project struct from a resource
func BuildStringFromProjectStruct(project *systemv3.Project) (string, error) {
	projectBytes, err := json.Marshal(&project)
	if err != nil {
		return "", err
	}

	return string(projectBytes), nil
}

// Build the project struct from a schema resource
func BuildProjectStructFromResource(d *schema.ResourceData) *systemv3.Project {

	projectStruct := systemv3.Project{
		Kind: "Project",
		Metadata: &commonv3.Metadata{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		},
	}

	if namespaceRoles, ok := d.GetOk("project_roles"); ok {
		projectStruct.Spec.ProjectNamespaceRoles = make([]*userv3.ProjectNamespaceRole, 0)
		rolesList := namespaceRoles.([]interface{})
		for _, eachRole := range rolesList {
			if role, ok := eachRole.(map[string]interface{}); ok {
				projectStruct.Spec.ProjectNamespaceRoles = append(projectStruct.Spec.ProjectNamespaceRoles, &userv3.ProjectNamespaceRole{
					Project: &projectStruct.Metadata.Name,
					Role:    role["role"].(string),
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
	project_roles := make([]map[string]interface{}, 0)
	for _, role := range project.Spec.GetProjectNamespaceRoles() {
		project_roles = append(project_roles, map[string]interface{}{
			"project":   role.Project,
			"role":      role.Role,
			"namespace": role.Namespace,
			"group":     role.Group,
		})
	}
	d.Set("project_roles", project_roles)
	user_roles := make([]map[string]interface{}, 0)
	for _, role := range project.Spec.GetUserRoles() {
		user_roles = append(user_roles, map[string]interface{}{
			"user": role.User,
			"role": role.Role,
		})
	}
	d.Set("user_roles", user_roles)
}
