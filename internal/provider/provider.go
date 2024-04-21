// Terraform provider
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rs "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iherbllc/terraform-provider-paralus/internal/datasources"
	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/resources"
)

var _ provider.Provider = (*paralusProvider)(nil)

type paralusProvider struct{}
type paralusProviderModel struct {
	Profile             types.String `tfsdk:"pctl_profile"`
	RestEndpoint        types.String `tfsdk:"pctl_rest_endpoint"`
	OPSEndpoint         types.String `tfsdk:"pctl_ops_endpoint"`
	APIKey              types.String `tfsdk:"pctl_api_key"`
	APISecret           types.String `tfsdk:"pctl_api_secret"`
	ConfigJSON          types.String `tfsdk:"pctl_config_json"`
	Partner             types.String `tfsdk:"pctl_partner"`
	Organization        types.String `tfsdk:"pctl_organization"`
	SkipServerCertValid types.String `tfsdk:"pctl_skip_server_cert_valid"`
}

func New() provider.Provider {
	return &paralusProvider{}
}

func (p *paralusProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "paralus"
}

func (p *paralusProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pctl_profile": rs.StringAttribute{
				Description: "PCTL Profile",
				Optional:    true,
			},
			"pctl_rest_endpoint": rs.StringAttribute{
				Description: "PCTL Profile",
				Optional:    true,
			},
			"pctl_ops_endpoint": rs.StringAttribute{
				Description: "OPS Endpoint",
				Optional:    true,
			},
			"pctl_api_key": rs.StringAttribute{
				Description: "PCTL API Key (obtained from UI). Either this and api_secret must be set config_json set",
				Optional:    true,
				Sensitive:   true,
			},
			"pctl_api_secret": rs.StringAttribute{
				Description: "PCTL API Secret (obtained from UI). Either this and api_key must be set config_json set",
				Optional:    true,
				Sensitive:   true,
			},
			"pctl_config_json": rs.StringAttribute{
				Description: "Config JSON (obtained from UI). Either this must be set or api_key/api_secret set",
				Optional:    true,
			},
			"pctl_partner": rs.StringAttribute{
				Optional: true,
			},
			"pctl_organization": rs.StringAttribute{
				Optional: true,
			},
			"pctl_skip_server_cert_valid": rs.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *paralusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config paralusProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile := os.Getenv("PCTL_PROFILE")
	rest_endpoint := os.Getenv("PCTL_REST_ENDPOINT")
	ops_endpoint := os.Getenv("PCTL_OPS_ENDPOINT")
	api_key := os.Getenv("PCTL_API_KEY")
	api_secret := os.Getenv("PCTL_API_SECRET")
	config_json := os.Getenv("PCTL_CONFIG_JSON")
	partner := os.Getenv("PCTL_PARTNER")
	organization := os.Getenv("PCTL_ORGANIZATION")
	skip_cert_valid := os.Getenv("PCTL_SKIP_SERVER_CERT_VALID")

	if !config.Profile.IsNull() {
		profile = config.Profile.ValueString()
	}
	if !config.RestEndpoint.IsNull() {
		rest_endpoint = config.RestEndpoint.ValueString()
	}
	if !config.OPSEndpoint.IsNull() {
		ops_endpoint = config.OPSEndpoint.ValueString()
	}
	if !config.APIKey.IsNull() {
		api_key = config.APIKey.ValueString()
	}
	if !config.APISecret.IsNull() {
		api_secret = config.APISecret.ValueString()
	}
	if !config.ConfigJSON.IsNull() {
		config_json = config.ConfigJSON.ValueString()
	}
	if !config.Partner.IsNull() {
		partner = config.Partner.ValueString()
	}
	if !config.Organization.IsNull() {
		organization = config.Organization.ValueString()
	}
	if !config.SkipServerCertValid.IsNull() {
		skip_cert_valid = config.SkipServerCertValid.ValueString()
	}

	if config_json == "" && api_key == "" && api_secret == "" && rest_endpoint == "" && ops_endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing necessary attributes",
			"Either pass in the individual API credentials and endpoints or a path to a json file containing those values",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := paralus.NewConfig(ctx, profile, rest_endpoint, ops_endpoint, api_key, api_secret, config_json, partner, organization, skip_cert_valid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to load configuration",
			"An unexpected error occurred while trying to create the paralus configuration. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Paralus Config Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *paralusProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &resources.ResourceCluster{}
		},
		func() resource.Resource {
			return &resources.ResourceProject{}
		},
		func() resource.Resource {
			return &resources.ResourceGroup{}
		},
	}
}

func (p *paralusProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return datasources.DataSourceBootstrapFile{}
		},
		func() datasource.DataSource {
			return datasources.DataSourceCluster{}
		},
		func() datasource.DataSource {
			return datasources.DataSourceProject{}
		},
		func() datasource.DataSource {
			return datasources.DataSourceGroup{}
		},
		func() datasource.DataSource {
			return datasources.DataSourceKubeConfig{}
		},
		func() datasource.DataSource {
			return datasources.DataSourceUsers{}
		},
	}
}

// // provider instance. Schema must either have all individual values set
// // or a path to a config file that can be loaded.
// func New() *schema.Provider {
// 	return &schema.Provider{
// 		Schema: map[string]*schema.Schema{},
// 		ResourcesMap: map[string]*schema.Resource{
// 			"paralus_cluster": resources.ResourceCluster(),
// 			"paralus_project": resources.ResourceProject(),
// 			"paralus_group":   resources.ResourceGroup(),
// 		},
// 		DataSourcesMap: map[string]*schema.Resource{
// 			"paralus_bootstrap_file": datasources.DataSourceBootstrapFile(),
// 			"paralus_cluster":        datasources.DataSourceCluster(),
// 			"paralus_project":        datasources.DataSourceProject(),
// 			"paralus_group":          datasources.DataSourceGroup(),
// 			"paralus_kubeconfig":     datasources.DataSourceKubeConfig(),
// 			"paralus_users":          datasources.DataSourceUsers(),
// 		},
// 		ConfigureContextFunc: providerConfigure,
// 	}
// }

// // Configure provider
// func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

// 	tflog.Debug(ctx, fmt.Sprintf(`Provider info:
// 	- pctl_profile: %s
// 	- pctl_rest_endpoint: %s
// 	- pctl_ops_endpoint: %s
// 	- pctl_config_json: %s
// 	- pctl_partner: %s
// 	- pctl_organization: %s
// 	- pctl_skip_server_cert_valid: %s
// 	`, d.Get("pctl_profile"),
// 		d.Get("pctl_rest_endpoint"), d.Get("pctl_ops_endpoint"),
// 		d.Get("pctl_config_json"), d.Get("pctl_partner"),
// 		d.Get("pctl_organization"), d.Get("pctl_skip_server_cert_valid")))

// 	return paralus.NewConfig(ctx, d)
// }
