package datasources

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

// / Paralus DataSource Cluster
func DataSourceProject() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a paralus project's information. Uses the [pctl](https://github.com/paralus/cli) library",
		ReadContext: datasourceProjectRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Project ID in the format \"PROJECT_NAME\"",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Project name",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Project description",
				Computed:    true,
			},
			"project_roles": {
				Type:        schema.TypeList,
				Description: "Project roles attached to project, containing group or namespace",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_roles": {
				Type:        schema.TypeList,
				Description: "User roles attached to project",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Retreive project JSON info
func datasourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId := d.Get("name").(string)

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	project, err := project.GetProjectByName(projectId)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error locating project %s",
			projectId)))
	}

	paralusUtils.BuildResourceFromProjectStruct(project, d)

	d.SetId(projectId)

	return diags

}
