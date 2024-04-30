// Project Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = (*DsProject)(nil)

func DataSourceProject() datasource.DataSource {
	return &DsProject{}
}

type DsProject struct {
	cfg *config.Config
}

func (d *DsProject) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Paralus DataSource Project
func (d *DsProject) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a paralus project's information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project ID in the format \"PROJECT_NAME\"",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Project description",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Project UUID",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"project_roles": schema.ListNestedBlock{
				MarkdownDescription: "Project roles attached to project, containing group or namespace",
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
			"user_roles": schema.ListNestedBlock{
				MarkdownDescription: "User roles attached to project",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.StringAttribute{
							Computed: true,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
						"namespace": schema.StringAttribute{
							MarkdownDescription: "Authorized namespace",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DsProject) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Retreive project JSON info
func (d *DsProject) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Project
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	projectId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("project name", projectId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	auth := d.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(d.cfg)))

	project, err := utils.GetProjectByName(ctx, projectId, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error locating project %s", projectId), err.Error())
		return
	}

	utils.BuildResourceFromProjectStruct(ctx, project, data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}
