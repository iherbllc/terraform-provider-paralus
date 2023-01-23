package paralus

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
)

// Generate a new proflie for making the call
func NewProfile(configJson string) *authprofile.Profile {
	return NewConfig(configJson).GetAppAuthProfile()
}

// Generates a new config
func NewConfig(configJson string) *config.Config {

	if configJson != "" {
		newConfig := config.GetConfig()
		if err := newConfig.Load(configJson); err != nil {
			tflog.Error(context.Background(), "Failed to load config from file.", map[string]interface{}{
				"config_json": configJson,
				"error":       err.Error(),
			})
		}
		return newConfig
	}

	return &config.Config{
		Profile:             os.Getenv("PCTL_PROFILE"),
		RESTEndpoint:        os.Getenv("PCTL_REST_ENDPOINT"),
		OPSEndpoint:         os.Getenv("PCTL_OPS_ENDPOINT"),
		APIKey:              os.Getenv("PCTL_API_KEY"),
		APISecret:           os.Getenv("PCTL_API_SECRET"),
		SkipServerCertValid: os.Getenv("PCTL_SKIP_SERVER_CERT_VALIDATION"),
	}
}
