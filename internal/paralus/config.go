// Client to retrieve new PCTL Config for connecting to paralus
package paralus

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/paralus/cli/pkg/config"
	"github.com/pkg/errors"
)

// Generates a new config either from a json file or via environment variables
func NewConfig(ctx context.Context, profile string, rest_endpoint string,
	ops_endpoint string, api_key string, api_secret string, config_json string, partner string,
	organization string, skip_cert_valid string) (*config.Config, error) {

	if config_json != "" {
		tflog.Debug(ctx, fmt.Sprintf("Using PCTL config json %s", config_json))
		newConfig, err := NewConfigFromFile(config_json)
		if err != nil {
			return nil, errors.Wrapf(err,
				"error parsing config_json file %s", config_json)
		}

		err = utils.AssertConfigNotEmpty(newConfig)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid loaded config %s", config_json))
		}

		return newConfig, nil
	}

	newConfig := config.GetConfig()
	newConfig.Profile = profile
	newConfig.RESTEndpoint = rest_endpoint
	newConfig.OPSEndpoint = ops_endpoint
	newConfig.APIKey = api_key
	newConfig.APISecret = api_secret
	newConfig.SkipServerCertValid = skip_cert_valid
	newConfig.Partner = partner
	newConfig.Organization = organization

	err := utils.AssertConfigNotEmpty(newConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error assigning config values")
	}

	return newConfig, nil
}

// Generate a new PCTL Config from a json path
func NewConfigFromFile(configJson string) (*config.Config, error) {
	newConfig := config.GetConfig()
	return newConfig, newConfig.Load(configJson)
}
