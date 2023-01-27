// Utility methods for Config struct
package utils

import (
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
