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
		ReadContext: datasourceProjectRead,
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

// Retreive project JSON info
func datasourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(d.Get("name").(string) + "ds")

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": d.Get("name").(string),
	})

	project, err := project.GetProjectByName(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error locating project %s",
			d.Get("name").(string))))
	}

	paralusUtils.BuildResourceFromProjectStruct(project, d)
	return diags

}
