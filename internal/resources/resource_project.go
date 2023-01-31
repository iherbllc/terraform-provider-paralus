// Project Terraform Resource
package resources

import (
	"context"
	"fmt"
	"strings"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/project"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// / Paralus Resource Project
func ResourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource containing paralus project information. Uses the [pctl](https://github.com/paralus/cli) library",
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Project ID in the format \"PROJECT_NAME\"",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Project name",
				ForceNew:    true,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Project description.",
				Optional:    true,
			},
			"uuid": {
				Type:        schema.TypeString,
				Description: "Project UUID",
				Computed:    true,
			},
			"project_roles": {
				Type:        schema.TypeList,
				Description: "Project roles attached to project, containing group or namespace",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							Description: "Project name",
							Required:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "Role name",
							Required:    true,
						},
						"namespace": {
							Type:        schema.TypeString,
							Description: "Authorized namespace",
							Optional:    true,
						},
						"group": {
							Type:        schema.TypeString,
							Description: "Authorized group",
							Optional:    true,
						},
					},
				},
			},
			"user_roles": {
				Type:        schema.TypeList,
				Description: "User roles attached to project",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:        schema.TypeString,
							Description: "Authorized user",
							Required:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "Authorized role",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Import an existing K8S cluster into a designated project
func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	projectId := d.Get("name").(string)

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(config.GetConfig())))

	diags := createOrUpdateProject(ctx, d, "POST")

	if diags.HasError() {
		return diags
	}

	d.SetId(projectId)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(config.GetConfig())))
	return createOrUpdateCluster(ctx, d, "PUT")
}

// Creates a new project or updates an existing one
func createOrUpdateProject(ctx context.Context, d *schema.ResourceData, requestType string) diag.Diagnostics {

	projectId := d.Get("name").(string)

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	tflog.Trace(ctx, fmt.Sprintf("Project %s request", requestType), map[string]interface{}{
		"project": projectId,
	})

	if requestType == "POST" {
		err := project.CreateProject(projectId, d.Get("description").(string))
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s project %s", howFail, projectId)))
		}
	} else if requestType == "PUT" {
		projectStruct := paralusUtils.BuildProjectStructFromResource(d)
		err := project.ApplyProject(projectStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s project %s", howFail,
					projectId)))
		}
	} else {
		return diag.FromErr(errors.Wrap(nil,
			fmt.Sprintf("Unknown request type %s", requestType)))
	}

	return resourceProjectRead(ctx, d, nil)
}

// Retreive project info
func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(config.GetConfig())))

	projectId := d.Get("name").(string)

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	projectStruct, err := project.GetProjectByName(projectId)
	if projectStruct == nil {
		// error should be "no rows in result set" but add it to TRACE in case it isn't.
		tflog.Trace(ctx, fmt.Sprintf("Error retrieving project info: %s", err))
		d.SetId("")
		return diags
	}

	// Update resource information from updated cluster
	paralusUtils.BuildResourceFromProjectStruct(projectStruct, d)

	return diags
}

// Import project into TF
func resourceProjectImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	projectId := d.Id()

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(config.GetConfig())))

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	projectStruct, err := project.GetProjectByName(projectId)
	// unlike others, fail and stop the import if we fail to get project info
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Project %s does not exist", projectId))
	}

	paralusUtils.BuildResourceFromProjectStruct(projectStruct, d)

	schemas := make([]*schema.ResourceData, 0)
	schemas = append(schemas, d)
	return schemas, nil

}

// Delete an existing cluster
func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(config.GetConfig())))

	projectId := d.Get("name").(string)

	tflog.Trace(ctx, "Deleting Project info", map[string]interface{}{
		"project": projectId,
	})

	err := project.DeleteProject(projectId)

	// The project doesn't exist
	if err != nil {
		// assume a no rows found error means the project does not exist
		if strings.Contains(fmt.Sprint(err), "no rows in result set") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed to delete project %s",
			projectId)))
	}

	d.SetId("")
	return diags
}
