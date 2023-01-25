// Client to retrieve new PCTL Config for connecting to paralus
package paralus

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/paralus/cli/pkg/config"
	"github.com/pkg/errors"
)

// Generates a new config either from a json file or via environment variables
func NewConfig(d *schema.ResourceData) (*config.Config, diag.Diagnostics) {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if configJson, ok := d.GetOk("pctl_config_Json"); ok {
		newConfig, err := NewConfigFromFile(configJson.(string))
		if err != nil {
			return nil, diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Error parsing config_json file %s", configJson)))
		}

		if newConfig.Organization == "" {
			return nil, diag.FromErr(errors.Wrap(nil, "organization missing from config_json file"))
		}

		return newConfig, diags
	}

	apiKey, ok := d.GetOk("pctl_api_key")
	if !ok {
		return nil, diag.FromErr(errors.Wrap(nil, "pctl_api_key must be set if config_json is not"))
	}

	apiSecret, ok := d.GetOk("pctl_api_secret")
	if !ok {
		return nil, diag.FromErr(errors.Wrap(nil, "pctl_api_secret must be set if config_json is not"))
	}

	return &config.Config{
		Profile:             d.Get("pctl_profile").(string),
		RESTEndpoint:        d.Get("pctl_rest_endpoint").(string),
		OPSEndpoint:         d.Get("pctl_ops_endpoint").(string),
		APIKey:              apiKey.(string),
		APISecret:           apiSecret.(string),
		SkipServerCertValid: d.Get("pctl_skip_server_cert_valid").(string),
		Partner:             d.Get("pctl_partner").(string),
		Organization:        d.Get("pctl_organization").(string),
	}, diags
}

// Generate a new PCTL Config from a json path
func NewConfigFromFile(configJson string) (*config.Config, error) {
	newConfig := config.GetConfig()
	return newConfig, newConfig.Load(configJson)
}
