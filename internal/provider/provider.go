package provider

import (
	"context"
	"os"

	"github.com/iherbllc/terraform-provider-paralus/internal/resources"
	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Instance() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_secret": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"cluster": resources.ResourceCluster(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	os.Setenv("PCTL_PROFILE", d.Get("paralus_profile").(string))
	os.Setenv("PCTL_REST_ENDPOINT", d.Get("paralus_rest_endpoint").(string))
	os.Setenv("PCTL_OPS_ENDPOINT", d.Get("paralus_ops_endpoit").(string))
	os.Setenv("PCTL_API_KEY", d.Get("api_key").(string))
	os.Setenv("PCTL_API_SECRET", d.Get("api_secret").(string))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return config.GetConfig().GetAppAuthProfile(), diags
}
