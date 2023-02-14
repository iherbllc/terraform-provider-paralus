// Utility methods for Config struct
package utils

import (
	"fmt"

	"github.com/paralus/cli/pkg/config"
)

// retrieve the Config as a map
func GetConfigAsMap(conf *config.Config) map[string]interface{} {
	configMap := map[string]interface{}{
		"partner":                conf.Partner,
		"rest_endpoint":          conf.RESTEndpoint,
		"ops_endpoint":           conf.OPSEndpoint,
		"organization":           conf.Organization,
		"profile":                conf.Profile,
		"skip_server_cert_valid": conf.SkipServerCertValid,
	}
	return configMap
}

func AssertConfigNotEmpty(conf *config.Config) error {
	if len(conf.Profile) == 0 {
		return fmt.Errorf("profile name not defined")
	}

	if len(conf.APIKey) == 0 {
		return fmt.Errorf("api key not defined")
	}

	if len(conf.APISecret) == 0 {
		return fmt.Errorf("api secret not defined")
	}

	if len(conf.Partner) == 0 {
		return fmt.Errorf("partner not defined")
	}
	if len(conf.RESTEndpoint) == 0 {
		return fmt.Errorf("rest endpoint not defined")
	}
	if len(conf.OPSEndpoint) == 0 {
		return fmt.Errorf("ops endpoint not defined")
	}

	return nil
}
