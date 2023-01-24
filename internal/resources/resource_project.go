package resources

import (
	"context"
	"fmt"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/project"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// / Paralus Resource Project
func ResourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"user_roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

// Import an existing K8S cluster into a designated project
func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	project := d.Get("name").(string)
	d.SetId(project)

	return createOrUpdateProject(ctx, d, "POST")
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateCluster(ctx, d, "PUT")
}

// Creates a new cluster or updates an existing one
func createOrUpdateProject(ctx context.Context, d *schema.ResourceData, requestType string) diag.Diagnostics {
	var diags diag.Diagnostics

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	tflog.Trace(ctx, fmt.Sprintf("Project %s request", requestType), map[string]interface{}{
		"project": d.Get("name").(string),
	})

	if requestType == "POST" {
		err := project.CreateProject(d.Get("name").(string), d.Get("description").(string))
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s project %s", howFail, d.Get("name"))))
		}
	} else if requestType == "PUT" {
		projectStruct := paralusUtils.BuildProjectStructFromResource(d)
		err := project.ApplyProject(projectStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s project %s", howFail,
					d.Get("name"))))
		}
	} else {
		return diag.FromErr(errors.Wrap(nil,
			fmt.Sprintf("Unknown request type %s", requestType)))
	}

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": d.Get("name").(string),
	})

	projectStruct, err := project.GetProjectByName(d.Get("name").(string))

	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to %s project %s", howFail,
				d.Get("name"))))
	}

	// Update resource information from updated cluster
	paralusUtils.BuildResourceFromProjectStruct(projectStruct, d)

	return diags
}

// Retreive project JSON info
func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": d.Get("name").(string),
	})

	_, err := project.GetProjectByName(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Project does not exist"))
	}

	return diags
}

// Delete an existing cluster
func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Trace(ctx, "Deleting Project info", map[string]interface{}{
		"project": d.Get("name").(string),
	})

	err := project.DeleteProject(d.Get("name").(string))

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed to delete project %s",
			d.Get("name").(string))))
	}

	return diags
}