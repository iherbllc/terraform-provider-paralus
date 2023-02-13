// Group Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Paralus DataSource Group
func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a paralus group's information. Uses the [pctl](https://github.com/paralus/cli) library",
		ReadContext: datasourceGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Group ID in the format \"GROUP_NAME\"",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Group name",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Group description",
				Computed:    true,
			},
			"project_roles": {
				Type:        schema.TypeList,
				Description: "Project roles attached to group, containing group or namespace",
				Computed:    true,
				Optional:    true,
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
			"users": {
				Type:        schema.TypeList,
				Description: "Users attached to group",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of group",
				Computed:    true,
			},
		},
	}
}

// Retreive group info
func datasourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	groupId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("group name", groupId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

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
	tflog.Debug(ctx, fmt.Sprintf("datasourceGroupRead provider config used: %s", utils.GetConfigAsMap(cfg)))

	group, err := utils.GetGroupByName(groupId, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "error locating group %s",
			groupId))
	}

	utils.BuildResourceFromGroupStruct(group, d)

	d.SetId(groupId)

	return diags

}
