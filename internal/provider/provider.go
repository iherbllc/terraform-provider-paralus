// Terraform provider
package provider

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/datasources"
	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/resources"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// provider instance. Schema must either have all individual values set
// or a path to a config file that can be loaded.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"pctl_profile": {
				Type:        schema.TypeString,
				Description: "PCTL Profile",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_PROFILE", nil),
			},
			"pctl_rest_endpoint": {
				Type:        schema.TypeString,
				Description: "Rest Endpoint",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_REST_ENDPOINT", nil),
			},
			"pctl_ops_endpoint": {
				Type:        schema.TypeString,
				Description: "OPS Endpoint",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_OPS_ENDPOINT", nil),
			},
			"pctl_api_key": {
				Type:        schema.TypeString,
				Description: "PCTL API Key (obtained from UI). Either this and api_secret must be set config_json set",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_API_KEY", nil),
			},
			"pctl_api_secret": {
				Type:        schema.TypeString,
				Description: "PCTL API Secret (obtained from UI). Either this and api_key must be set config_json set",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_API_SECRET", nil),
			},
			"pctl_config_json": {
				Type:        schema.TypeString,
				Description: "Config JSON (obtained from UI). Either this must be set or api_key/api_secret set",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_CONFIG_JSON", nil),
			},
			"pctl_partner": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_PARTNER", nil),
			},
			"pctl_organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_ORGANIZATION", nil),
			},
			"pctl_skip_server_cert_valid": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PCTL_SKIP_SERVER_CERT_VALID", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"paralus_cluster": resources.ResourceCluster(),
			"paralus_project": resources.ResourceProject(),
			"paralus_group":   resources.ResourceGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"paralus_bootstrap_file": datasources.DataSourceBootstrapFile(),
			"paralus_cluster":        datasources.DataSourceCluster(),
			"paralus_project":        datasources.DataSourceProject(),
			"paralus_group":          datasources.DataSourceGroup(),
			"paralus_kubeconfig":     datasources.DataSourceKubeConfig(),
			"paralus_users":          datasources.DataSourceUsers(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// Configure provider
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	tflog.Debug(ctx, fmt.Sprintf(`Provider info:
	- pctl_profile: %s
	- pctl_rest_endpoint: %s
	- pctl_ops_endpoint: %s
	- pctl_config_json: %s
	- pctl_partner: %s
	- pctl_organization: %s
	- pctl_skip_server_cert_valid: %s
	`, d.Get("pctl_profile"),
		d.Get("pctl_rest_endpoint"), d.Get("pctl_ops_endpoint"),
		d.Get("pctl_config_json"), d.Get("pctl_partner"),
		d.Get("pctl_organization"), d.Get("pctl_skip_server_cert_valid")))

	return paralus.NewConfig(ctx, d)
}
