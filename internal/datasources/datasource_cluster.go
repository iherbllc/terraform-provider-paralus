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

var _ datasource.DataSource = (*DsCluster)(nil)

func DataSourceCluster() datasource.DataSource {
	return &DsCluster{}
}

type DsCluster struct {
	cfg *config.Config
}

func (d *DsCluster) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Paralus DataSource Cluster
func (d *DsCluster) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a paralus cluster's information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID in the format \"PROJECT_NAME:CLUSTER_NAME\"",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Cluster name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Cluster description",
				Computed:            true,
			},
			"cluster_type": schema.StringAttribute{
				MarkdownDescription: "Cluster type. For example, \"imported.\" ",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Cluster UUID",
				Computed:            true,
			},
			"project": schema.StringAttribute{
				MarkdownDescription: "Project containing cluster",
				Required:            true,
			},
			// Will only ever be updated by provider
			"bootstrap_files_combined": schema.StringAttribute{
				MarkdownDescription: "YAML files used to deploy paralus agent to the cluster stored as a single massive file",
				Computed:            true,
			},
			// Will only ever be updated by provider
			"bootstrap_files": schema.ListAttribute{
				MarkdownDescription: "YAML files used to deploy paralus agent to the cluster stored as a list",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Map of lables to include for cluster",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"annotations": schema.MapAttribute{
				MarkdownDescription: "Map of annotations to include for cluster",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"relays": schema.StringAttribute{
				MarkdownDescription: "Relays information",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"params": schema.SingleNestedBlock{
				MarkdownDescription: "Import parameters",
				Attributes: map[string]schema.Attribute{
					"provision_type": schema.StringAttribute{
						MarkdownDescription: "Provision Type. For example, \"IMPORT\"",
						Computed:            true,
					},
					"provision_environment": schema.StringAttribute{
						MarkdownDescription: "Provision Environment. For example, \"CLOUD\"",
						Computed:            true,
					},
					"provision_package_type": schema.StringAttribute{
						MarkdownDescription: "Provision Type. For example, \"LINUX\"",
						Computed:            true,
					},
					"environment_provider": schema.StringAttribute{
						MarkdownDescription: "Provision Type. For example, \"GCP\"",
						Computed:            true,
					},
					"kubernetes_provider": schema.StringAttribute{
						MarkdownDescription: "Provision Type. For example, \"EKS\"",
						Computed:            true,
					},
					"state": schema.StringAttribute{
						MarkdownDescription: "Provision Type. For example, \"PROVISION\"",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *DsCluster) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Retreive cluster JSON info
func (d *DsCluster) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	// Prevent panic if the provider has not been configured.
	if d.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Cluster
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	auth := d.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(d.cfg)))

	clusterStruct, err := utils.GetCluster(ctx, clusterId, projectId, auth)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error locating cluster %s in project %s",
			clusterId, projectId), err.Error())
		return
	}

	diags = utils.BuildResourceFromClusterStruct(ctx, clusterStruct, data, auth)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}
