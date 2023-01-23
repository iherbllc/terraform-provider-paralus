package provider

import (
	"context"
	"os"

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
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"rest_endpoint": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"ops_endpoint": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"api_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"api_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"config_json": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"partner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"paralus_cluster": resources.ResourceCluster(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"paralus_cluster": datasources.DataSourceCluster(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	os.Setenv("PCTL_PROFILE", d.Get("profile").(string))
	os.Setenv("PCTL_REST_ENDPOINT", d.Get("rest_endpoint").(string))
	os.Setenv("PCTL_OPS_ENDPOINT", d.Get("ops_endpoint").(string))
	os.Setenv("PCTL_API_KEY", d.Get("api_key").(string))
	os.Setenv("PCTL_API_SECRET", d.Get("api_secret").(string))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return paralus.NewProfile(d.Get("config_json").(string)), diags
}
