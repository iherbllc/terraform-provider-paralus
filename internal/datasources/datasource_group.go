// Group Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dsGroup struct {
}

func (d *dsGroup) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Paralus DataSource Cluster
func (d *dsGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a paralus group's information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Group ID in the format \"GROUP_NAME\"",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Group name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Group description",
				Computed:    true,
			},
			"project_roles": schema.ListNestedAttribute{
				Description: "Project roles attached to group, containing group or namespace",
				Computed:    true,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project": schema.StringAttribute{
							Computed: true,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
						"namespace": schema.StringAttribute{
							Computed: true,
						},
						"group": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"users": schema.ListAttribute{
				Description: "Users attached to group",
				Computed:    true,
				ElementType: types.StringType,
			},
			"type": schema.StringAttribute{
				Description: "Type of group",
				Computed:    true,
			},
		},
	}
}

// Retreive group info
func (d *dsGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data *structs.Group

	groupId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("group name", groupId)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	var cfg *config.Config
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(cfg)))

	group, err := utils.GetGroupByName(ctx, groupId, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error locating group %s",
			groupId), err.Error())
		return
	}

	utils.BuildResourceFromGroupStruct(ctx, group, data)

	if resp.Diagnostics.HasError() {
		data.Id = types.StringNull()
		return
	}

	data.Id = types.StringValue(groupId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}
