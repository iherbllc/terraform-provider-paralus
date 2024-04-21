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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dsProject struct {
}

func (d *dsProject) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Paralus DataSource Project
func (d *dsProject) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a paralus project's information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Project ID in the format \"PROJECT_NAME\"",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Project name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Project description",
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "Project UUID",
				Computed:    true,
			},
			"project_roles": schema.ListNestedAttribute{
				Description: "Project roles attached to project, containing group or namespace",
				Computed:    true,
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
			"user_roles": schema.ListNestedAttribute{
				Description: "User roles attached to project",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.StringAttribute{
							Computed: true,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
						"namespace": schema.StringAttribute{
							Description: "Authorized namespace",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Retreive project JSON info
func (d *dsProject) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data *structs.Project

	projectId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("project name", projectId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	var cfg *config.Config
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(cfg)))

	project, err := utils.GetProjectByName(ctx, projectId, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error locating project %s", projectId), err.Error())
		return
	}

	utils.BuildResourceFromProjectStruct(ctx, project, data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}
