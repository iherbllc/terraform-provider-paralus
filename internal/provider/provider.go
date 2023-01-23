package provider

import (
	"context"

	"github.com/iherbllc/terraform-provider-paralus/internal/datasources"
	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/resources"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// provider instance. Schema must either have all individual values set
// or a path to a config file that can be loaded.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"profile": {
				Type:        schema.TypeString,
				Description: "PCTL Profile",
				Optional:    true,
				Sensitive:   true,
			},
			"rest_endpoint": {
				Type:        schema.TypeString,
				Description: "Rest Endpoint",
				Optional:    true,
				Sensitive:   true,
			},
			"ops_endpoint": {
				Type:        schema.TypeString,
				Description: "OPS Endpoint",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key": {
				Type:        schema.TypeString,
				Description: "PCTL API Key (obtained from UI). Either this and api_secret must be set config_json set",
				Optional:    true,
				Sensitive:   true,
			},
			"api_secret": {
				Type:        schema.TypeString,
				Description: "PCTL API Secret (obtained from UI). Either this and api_key must be set config_json set",
				Optional:    true,
				Sensitive:   true,
			},
			"config_json": {
				Type:        schema.TypeString,
				Description: "Config JSON (obtained from UI). Either this must be set or api_key/api_secret set",
				Optional:    true,
			},
			"partner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"skip_server_cert_valid": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"paralus_cluster": resources.ResourceCluster(),
			"paralus_project": resources.ResourceProject(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"paralus_cluster": datasources.DataSourceCluster(),
			"paralus_project": datasources.DataSourceProject(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// Configure provider
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	return paralus.NewConfig(d)
}
