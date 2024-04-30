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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = (*DsGroup)(nil)

func DataSourceGroup() datasource.DataSource {
	return &DsGroup{}
}

type DsGroup struct {
	cfg *config.Config
}

func (d *DsGroup) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Paralus DataSource Cluster
func (d *DsGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a paralus group's information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Group ID in the format \"GROUP_NAME\"",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Group name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Group description",
				Computed:            true,
			},
			"users": schema.ListAttribute{
				MarkdownDescription: "Users attached to group",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of group",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"project_roles": schema.ListNestedBlock{
				MarkdownDescription: "Project roles attached to group, containing group or namespace",
				NestedObject: schema.NestedBlockObject{
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
		},
	}
}

func (d *DsGroup) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.cfg = cfg
}

// Retreive group info
func (d *DsGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Group
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("group name", groupId)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	auth := d.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(d.cfg)))

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
