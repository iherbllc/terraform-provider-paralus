package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/rerror"
)

// Looks directly for a project based on info provided
func GetProjectFast(ctx context.Context, auth *authprofile.Profile,
	partner string, organization string, project string) (string, error) {

	if auth == nil {
		auth = config.GetConfig().GetAppAuthProfile()
	}

	uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/project/%s", partner, organization, project)

	tflog.Trace(ctx, "Project GET API Request", map[string]interface{}{
		"uri":    uri,
		"method": "GET",
	})

	return auth.AuthAndRequest(uri, "GET", nil)

}

// retrieve all projects from paralus
func ListAllProjects(ctx context.Context, auth *authprofile.Profile, partner string, organization string) ([]*systemv3.Project, error) {
	var projects []*systemv3.Project
	limit := 10000
	c, count, err := listProjects(ctx, auth, partner, organization, limit, 0)
	if err != nil {
		return nil, err
	}
	projects = c
	for count > limit {
		offset := limit
		limit = count
		c, _, err = listProjects(ctx, auth, partner, organization, limit, offset)
		if err != nil {
			return projects, err
		}
		projects = append(projects, c...)
	}
	return projects, nil
}

// build a list of all projects
func listProjects(ctx context.Context, auth *authprofile.Profile,
	partner string, organization string, limit, offset int) ([]*systemv3.Project, int, error) {
	// check to make sure the limit or offset is not negative
	if limit < 0 || offset < 0 {
		return nil, 0, fmt.Errorf("provided limit (%d) or offset (%d) cannot be negative", limit, offset)
	}

	uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/projects", partner, organization, limit, offset)

	tflog.Trace(ctx, "All Projects GET API Request", map[string]interface{}{
		"uri":    uri,
		"method": "GET",
	})

	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return nil, 0, rerror.CrudErr{
			Type: "project",
			Name: "",
			Op:   "list",
		}
	}

	resp_interf, err := JsonToMap(resp)

	if err != nil {
		return nil, 0, err
	}

	tflog.Trace(ctx, "All Projects GET API Request", resp_interf)

	a := systemv3.ProjectList{}

	if err := json.Unmarshal([]byte(resp), &a); err != nil {
		return nil, 0, err
	}

	return a.Items, int(a.Metadata.Count), nil
}

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
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			Organization: d.Get("organization").(string),
			Partner:      d.Get("partner").(string),
		},
		Spec: &userv3.ProjectNamespaceRole[],
	}

	return &projectStruct
}

// Build a resource from a project struct
func BuildResourceFromProjectString(project string, d *schema.ResourceData) error {
	// Need to take json project and convert to the new version
	projectBytes := []byte(project)
	projectStruct := infrav3.ProjectCluster{}
	if err := json.Unmarshal(projectBytes, &projectStruct); err != nil {
		return err
	}

	BuildResourceFromProjectStruct(&projectStruct, d)

	return nil
}

// Build the schema resource from project Struct
func BuildResourceFromProjectStruct(project *systemv3.Project, d *schema.ResourceData) {
	d.Set("name", project.Metadata.Name)
	d.Set("description", project.Metadata.Description)
	d.Set("organization", project.Metadata.Organization)
	d.Set("partner", project.Metadata.Partner)
	d.Set("project_namespace_roles", project.Spec.ProjectNamespaceRoles)
}
