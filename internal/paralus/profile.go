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

	if configJson, ok := d.GetOk("config_Json"); ok {
		newConfig, err := NewConfigFromFile(configJson.(string))
		if err != nil {
			return nil, diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Error parsing config_json file %s", configJson)))
		}
		return newConfig, diags
	}

	apiKey, ok := d.GetOk("api_key")
	if !ok {
		return nil, diag.FromErr(errors.Wrap(nil, "api_key must be set if config_json is not"))
	}

	apiSecret, ok := d.GetOk("api_secret")
	if !ok {
		return nil, diag.FromErr(errors.Wrap(nil, "api_secret must be set if config_json is not"))
	}

	return &config.Config{
		Profile:             d.Get("profile").(string),
		RESTEndpoint:        d.Get("rest_endpoint").(string),
		OPSEndpoint:         d.Get("ops_endpoint").(string),
		APIKey:              apiKey.(string),
		APISecret:           apiSecret.(string),
		SkipServerCertValid: d.Get("skip_server_cert_valid").(string),
		Partner:             d.Get("partner").(string),
		Organization:        d.Get("organization").(string),
	}, diags
}

// Generate a new PCTL Config from a json path
func NewConfigFromFile(configJson string) (*config.Config, error) {
	newConfig := config.GetConfig()
	return newConfig, newConfig.Load(configJson)
}
