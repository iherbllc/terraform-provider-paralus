// Cluster Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/paralus/cli/pkg/config"
)

type dsBSFile struct {
}

func (d *dsBSFile) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bootstrap_file"
}

func (d *dsBSFile) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the bootstrap file generated after a cluster is imported. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Cluster ID in the format \"PROJECT_NAME:CLUSTER_NAME\"",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Cluster name",
				Required:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "Cluster UUID",
				Computed:    true,
			},
			"project": schema.StringAttribute{
				Description: "Project containing cluster",
				Required:    true,
			},
			// Will only ever be updated by provider
			"bootstrap_files_combined": schema.StringAttribute{
				Description: "YAML files used to deploy paralus agent to the cluster stored as a single massive file",
				Computed:    true,
			},
			// Will only ever be updated by provider
			"bootstrap_files": schema.ListAttribute{
				Description: "YAML files used to deploy paralus agent to the cluster stored as a list",
				Computed:    true,
				ElementType: types.StringType,
			},
			"relays": schema.StringAttribute{
				Description: "Relays information",
				Computed:    true,
			},
		},
	}
}

// Retreive cluster bootstrap file
func (d *dsBSFile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *structs.BootstrapFileData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	projectId := data.Project.ValueString()
	clusterId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	resp.Diagnostics.Append(diags...)
	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(clusterId + ":" + projectId)

	tflog.Trace(ctx, "Retrieving bootstrap info", map[string]interface{}{
		"cluster": clusterId,
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

	clusterStruct, err := utils.GetCluster(ctx, clusterId, projectId, auth)

	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError(fmt.Sprintf("error locating cluster %s in project %s",
			clusterId, projectId), err.Error())
		return
	}

	err, relays, bsfiles, bsfile := utils.SetBootstrapFileAndRelays(ctx, projectId, clusterId, auth)
	if err != nil {
		resp.Diagnostics.AddError("Setting bootstrap file and relays failed", err.Error())
		return
	}
	data.Relays = types.StringValue(relays)
	data.BSFiles, diags = types.ListValueFrom(ctx, types.StringType, bsfiles)
	resp.Diagnostics.Append(diags...)
	data.BSFileCombined = types.StringValue(bsfile)
	data.Uuid = types.StringValue(clusterStruct.Metadata.Id)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
