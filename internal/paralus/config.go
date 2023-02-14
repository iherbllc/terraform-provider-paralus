// Client to retrieve new PCTL Config for connecting to paralus
package paralus

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/paralus/cli/pkg/config"
	"github.com/pkg/errors"
)

// Generates a new config either from a json file or via environment variables
func NewConfig(ctx context.Context, d *schema.ResourceData) (*config.Config, diag.Diagnostics) {

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if configJson, ok := d.GetOk("pctl_config_json"); ok {
		tflog.Debug(ctx, fmt.Sprintf("Using PCTL config json %s", configJson))
		newConfig, err := NewConfigFromFile(configJson.(string))
		if err != nil {
			return nil, diag.FromErr(errors.Wrapf(err,
				"error parsing config_json file %s", configJson))
		}

		err = utils.AssertConfigNotEmpty(newConfig)
		if err != nil {
			return nil, diag.FromErr(errors.Wrap(err, fmt.Sprintf("invalid loaded config %s", configJson)))
		}

		return newConfig, diags
	}

	newConfig := config.GetConfig()
	newConfig.Profile = d.Get("pctl_profile").(string)
	newConfig.RESTEndpoint = d.Get("pctl_rest_endpoint").(string)
	newConfig.OPSEndpoint = d.Get("pctl_ops_endpoint").(string)
	newConfig.APIKey = d.Get("pctl_api_key").(string)
	newConfig.APISecret = d.Get("pctl_api_secret").(string)
	newConfig.SkipServerCertValid = d.Get("pctl_skip_server_cert_valid").(string)
	newConfig.Partner = d.Get("pctl_partner").(string)
	newConfig.Organization = d.Get("pctl_organization").(string)

	err := utils.AssertConfigNotEmpty(newConfig)
	if err != nil {
		return nil, diag.FromErr(errors.Wrap(err, "error assigning config values"))
	}

	return newConfig, diags
}

// Generate a new PCTL Config from a json path
func NewConfigFromFile(configJson string) (*config.Config, error) {
	newConfig := config.GetConfig()
	return newConfig, newConfig.Load(configJson)
}
