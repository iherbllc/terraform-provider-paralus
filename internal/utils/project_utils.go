// Utility methods for PCTL Project struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
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
					Project:   &projectStruct.Metadata.Name, // project will always default to the project name to avoid user error
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
					User:      role["user"].(string),
					Role:      role["role"].(string),
					Namespace: role["namespace"].(string),
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
	for _, role := range project.Spec.UserRoles {
		userRoles = append(userRoles, map[string]interface{}{
			"user":      role.User,
			"role":      role.Role,
			"namespace": role.Namespace,
		})
	}
	d.Set("user_roles", userRoles)
}

// Check to make sure that the list of roles from ProjectNamespaceRole has unique role values.
// This is required due to a limitation with Paralus.
// See: https://github.com/paralus/paralus/issues/136
func AssertUniqueRoles(pnrStruct []*userv3.ProjectNamespaceRole) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(pnrStruct) >= 2 {
		pnrStructMap := make(map[string]string)
		for _, role := range pnrStruct {
			if _, exists := pnrStructMap[role.Role]; exists {
				return diag.FromErr(errors.New("roles must be distinct between project_roles blocks. If the same is required, then grant through the group instead"))
			}
			pnrStructMap[role.Role] = "unique"
		}

	}

	return diags
}

// Assures that the combination of project, role, namespace, and group are all unique
func AssertUniquePRNStruct(pnrStruct []*userv3.ProjectNamespaceRole) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(pnrStruct) >= 2 {
		pnrStructMap := make(map[string]string)
		for _, role := range pnrStruct {
			pnrkey := fmt.Sprintf("%s,%s,%s,%s", *role.Group, *role.Namespace, *role.Project, role.Role)
			if _, exists := pnrStructMap[pnrkey]; exists {
				return diag.FromErr(fmt.Errorf("group, namespace, project, and role entry already found: '%s'. must have a unique combination", pnrkey))
			}
			pnrStructMap[pnrkey] = "unique"
		}

	}

	return diags
}

// Check projects specified in the ProjectNamespaceRoles struct exist in Paralus
func CheckProjectsFromPNRStructExist(ctx context.Context, pnrStruct []*userv3.ProjectNamespaceRole, auth *authprofile.Profile) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(pnrStruct) > 0 {
		for _, pnr := range pnrStruct {
			projectName := pnr.Project
			if projectName != nil {
				// if we get an empty project name, verify the role allows it
				if *projectName == "" {
					diags = CheckAllowEmptyProject(pnr.Role)
					if diags.HasError() {
						return diags
					}
					continue
				}
				_, err := GetProjectByName(ctx, *projectName, auth)
				if err != nil {
					if err == ErrResourceNotExists {
						return diag.FromErr(fmt.Errorf("project '%s' does not exist", *projectName))
					}
					return diag.FromErr(errors.Wrapf(err, "error getting project %s info",
						*projectName))
				}

			}
		}
	}

	return diags
}

// Thesea are the roles that don't require specifying a project
var NON_PROJECT_ROLES = []string{"ADMIN", "ADMIN_READ_ONLY"}

// Check the role desired allows for no project specified
func CheckAllowEmptyProject(role string) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, nonProjectRole := range NON_PROJECT_ROLES {
		if nonProjectRole == role {
			return diags
		}
	}

	return diag.FromErr(fmt.Errorf("project must be specified when assigning role '%s'", role))
}

// Get project by name
func GetProjectByName(ctx context.Context, projectName string, auth *authprofile.Profile) (*systemv3.Project, error) {
	cfg := config.GetConfig()
	uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/project/%s", cfg.Partner, cfg.Organization, projectName)
	resp, err := makeRestCall(ctx, uri, "GET", nil, auth)
	if err != nil {
		return nil, err
	}
	proj := &systemv3.Project{}
	err = json.Unmarshal([]byte(resp), proj)
	if err != nil {
		return nil, err
	}

	return proj, nil
}

// Apply project takes the project details and sends it to the core
func ApplyProject(ctx context.Context, proj *systemv3.Project, auth *authprofile.Profile) error {
	cfg := config.GetConfig()
	projExisting, err := GetProjectByName(ctx, proj.Metadata.Name, auth)
	if projExisting != nil {
		tflog.Debug(context.Background(), fmt.Sprintf("updating project: %s", proj.Metadata.Name))
		uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/project/%s", cfg.Partner, cfg.Organization, proj.Metadata.Name)
		_, err := makeRestCall(ctx, uri, "PUT", proj, auth)
		if err != nil {
			return err
		}
	} else {
		if err != nil && err != ErrResourceNotExists {
			return err
		}
		tflog.Debug(context.Background(), fmt.Sprintf("creating project: %s", proj.Metadata.Name))
		uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/project", cfg.Partner, cfg.Organization)
		_, err := makeRestCall(ctx, uri, "POST", proj, auth)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete project
func DeleteProject(ctx context.Context, project string, auth *authprofile.Profile) error {
	cfg := config.GetConfig()

	// Need to add delay to cluster list check due to the circumstances
	// where the cluster resource deletion is faster then the API process to remove the clusters.
	// Using jitter for better randomness
	b := &backoff.Backoff{
		Max:    5 * time.Minute,
		Jitter: true,
	}

	rand.Seed(42)

	for {
		// Before delete, let's make sure the project is empty
		clusters, err := ListAllClusters(ctx, project, auth)
		if len(clusters) == 0 || err == ErrResourceNotExists {
			uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/project/%s", cfg.Partner, cfg.Organization, project)
			_, err := makeRestCall(ctx, uri, "DELETE", nil, auth)
			return err
		}
		if err != nil && err != ErrResourceNotExists {
			return err
		}

		d := b.Duration()
		if d >= b.Max {
			break
		}
		tflog.Info(context.Background(), fmt.Sprintf("Project %s is not empty. Will check again in %s", project, d))
		time.Sleep(d)

	}
	maxCheck := b.Duration()
	b.Reset()
	return fmt.Errorf("after %s, project %s is still not empty. Remove all clusters before attempting deletion",
		maxCheck, project)

}
