package paralus

import (
	"os"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
)

// Generate a new proflie for making the call
func NewProfile() *authprofile.Profile {
	return NewConfig().GetAppAuthProfile()
}

// Generates a new config
func NewConfig() *config.Config {
	return &config.Config{
		Profile:             os.Getenv("PCTL_PROFILE"),
		RESTEndpoint:        os.Getenv("PCTL_REST_ENDPOINT"),
		OPSEndpoint:         os.Getenv("PCTL_OPS_ENDPOINT"),
		APIKey:              os.Getenv("PCTL_API_KEY"),
		APISecret:           os.Getenv("PCTL_API_SECRET"),
		SkipServerCertValid: os.Getenv("PCTL_SKIP_SERVER_CERT_VALIDATION"),
	}
}
