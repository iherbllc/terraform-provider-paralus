// KubeConfig Terraform DataSource
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

var _ datasource.DataSource = (*DsKubeConfig)(nil)

func DataSourceKubeConfig() datasource.DataSource {
	return &DsKubeConfig{}
}

type DsKubeConfig struct {
	cfg *config.Config
}

func (d *DsKubeConfig) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubeconfig"
}

// Paralus DataSource Cluster
func (d *DsKubeConfig) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a user's kubeconfig information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of user to retrieve kubeconfig of. Note: User must have already generate a kubeconfig file at least once through the UI to be able to retrieve it",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Namespace to set as the default for the kubeconfig",
				Optional:            true,
			},
			"cluster": schema.StringAttribute{
				MarkdownDescription: "Cluster to get certificate information for",
				Optional:            true,
			},
			"cluster_info": schema.ListNestedAttribute{
				MarkdownDescription: "KubeConfig cluster information",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"certificate_authority_data": schema.StringAttribute{
							MarkdownDescription: "Certificate authority data for cluster",
							Computed:            true,
							Sensitive:           true,
						},
						"server": schema.StringAttribute{
							MarkdownDescription: "URL to server",
							Computed:            true,
						},
					},
				},
			},
			"client_certificate_data": schema.StringAttribute{
				MarkdownDescription: "Client certificate data",
				Computed:            true,
				Sensitive:           true,
			},
			"client_key_data": schema.StringAttribute{
				MarkdownDescription: "Client key data",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *DsKubeConfig) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Retreive KubeConfig JSON info
func (d *DsKubeConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var diags diag.Diagnostics
	var data *structs.KubeConfig

	userName := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("name", userName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving kubeconfig info", map[string]interface{}{
		"name": userName,
	})

	cluster := data.Cluster.ValueString()
	namespace := data.Namespace.ValueString()

	auth := d.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(d.cfg)))
	userInfo, err := utils.GetUserByName(ctx, userName, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error locating user info: %s", userName), err.Error())
		return
	}

	userID := userInfo.Metadata.Id

	tflog.Trace(ctx, "Retrieving KubeConfig info", map[string]interface{}{
		"name":      userName,
		"id":        userID,
		"cluster":   cluster,
		"namespace": namespace,
	})

	kubeConfig, err := utils.GetKubeConfig(ctx, userID, namespace, cluster, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error locating kubeconfig for user %s. Make sure "+
			"the kubeconfig has been generated manually through the UI for the first time.", userName), err.Error())
	}

	utils.BuildKubeConfigStruct(ctx, data, kubeConfig)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}
