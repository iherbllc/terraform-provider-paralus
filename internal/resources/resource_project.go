// Project Terraform Resource
package resources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Paralus Resource Project
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
							Description: "Project name. This will always be the same as the resource project name.",
							Computed:    true,
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
							Required:    true,
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
						"namespace": {
							Type:        schema.TypeString,
							Description: "Authorized namespace",
							Optional:    true,
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

	var cfg *config.Config
	if m == nil {
		cfg = config.GetConfig()
	} else {
		cfg = m.(*config.Config)
		if cfg == nil {
			cfg = config.GetConfig()
		}
	}
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectCreate provider config used: %s", utils.GetConfigAsMap(cfg)))
	diags := createOrUpdateProject(ctx, d, "POST", auth)

	if diags.HasError() {
		return diags
	}

	d.SetId(projectId)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var cfg *config.Config
	if m == nil {
		cfg = config.GetConfig()
	} else {
		cfg = m.(*config.Config)
		if cfg == nil {
			cfg = config.GetConfig()
		}
	}
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectUpdate provider config used: %s", utils.GetConfigAsMap(cfg)))

	return createOrUpdateProject(ctx, d, "PUT", auth)
}

// Creates a new project or updates an existing one
func createOrUpdateProject(ctx context.Context, d *schema.ResourceData, requestType string, auth *authprofile.Profile) diag.Diagnostics {

	projectId := d.Get("name").(string)

	diags := utils.AssertStringNotEmpty("project name", projectId)
	if diags.HasError() {
		return diags
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	tflog.Trace(ctx, fmt.Sprintf("Project %s request", requestType), map[string]interface{}{
		"project": projectId,
	})

	projectStruct := utils.BuildProjectStructFromResource(d)

	// before creating the project, verify that requested group exists
	diags = utils.CheckGroupsFromPNRStructExist(projectStruct.Spec.GetProjectNamespaceRoles(), auth)
	if diags.HasError() {
		return diags
	}

	// before creating the project, verify that users in question exist
	diags = utils.CheckUserRoleUsersExist(projectStruct.Spec.GetUserRoles(), auth)
	if diags.HasError() {
		return diags
	}

	// Required due to limitation with paralus
	// See issue: https://github.com/paralus/paralus/issues/136
	// Remove once paralus supports this feature.
	diags = utils.AssertUniqueRoles(projectStruct.Spec.ProjectNamespaceRoles)
	if diags.HasError() {
		return diags
	}

	err := utils.ApplyProject(projectStruct, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err,
			"failed to %s project %s", howFail,
			projectId))
	}

	return resourceProjectRead(ctx, d, nil)
}

// Retreive project info
func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	var cfg *config.Config
	if m == nil {
		cfg = config.GetConfig()
	} else {
		cfg = m.(*config.Config)
		if cfg == nil {
			cfg = config.GetConfig()
		}
	}
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectRead provider config used: %s", utils.GetConfigAsMap(cfg)))

	projectId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("project name", projectId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	projectStruct, err := utils.GetProjectByName(projectId, auth)
	if err == utils.ErrResourceNotExists {
		d.SetId("")
		return diags
	}
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Error retrieving info for project %s", projectId))
	}

	// Update resource information from updated cluster
	utils.BuildResourceFromProjectStruct(projectStruct, d)

	return diags
}

// Import project into TF
func resourceProjectImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	projectId := d.Id()

	var cfg *config.Config
	if m == nil {
		cfg = config.GetConfig()
	} else {
		cfg = m.(*config.Config)
		if cfg == nil {
			cfg = config.GetConfig()
		}
	}
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectImport provider config used: %s", utils.GetConfigAsMap(cfg)))

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	projectStruct, err := utils.GetProjectByName(projectId, auth)
	// unlike others, fail and stop the import if we fail to get project info
	if err != nil {
		return nil, errors.Wrapf(err, "project %s does not exist", projectId)
	}

	utils.BuildResourceFromProjectStruct(projectStruct, d)

	schemas := make([]*schema.ResourceData, 0)
	schemas = append(schemas, d)
	return schemas, nil

}

// Delete an existing cluster
func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	var cfg *config.Config
	if m == nil {
		cfg = config.GetConfig()
	} else {
		cfg = m.(*config.Config)
		if cfg == nil {
			cfg = config.GetConfig()
		}
	}
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectDelete provider config used: %s", utils.GetConfigAsMap(cfg)))

	projectId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("project name", projectId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Deleting Project info", map[string]interface{}{
		"project": projectId,
	})

	// verify project exists before attempting delete
	_, err := utils.GetProjectByName(projectId, auth)
	if err != nil && err != utils.ErrResourceNotExists {
		return diag.FromErr(errors.Wrapf(err, "failed to retrieve project %s",
			projectId))
	}

	err = utils.DeleteProject(projectId, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "failed to delete project %s",
			projectId))
	}

	d.SetId("")
	return diags
}
