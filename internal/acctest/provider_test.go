package acctest

import (
	"fmt"
	"testing"

	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/provider"
	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

// // credential ENV vars
// var credsEnvVars = []string{
// 	"PCTL_API_KEY",
// 	"PCTL_API_SECRET",
// }

// // endpoint ENV vars
// var endpointEnvVars = []string{
// 	"PCTL_REST_ENDPOINT",
// 	"PCTL_OPS_ENDPOINT",
// }

// var configEnvVars = []string{
// 	"PCTL_PROFILE",
// }

func init() {
	testAccProvider = provider.Provider()
	testAccProviders = map[string]*schema.Provider{
		"paralusctl": testAccProvider,
	}
}

// Return test provider config
func paralusProviderConfig() *config.Config {
	// utils.LoadEnv()
	return paralus.NewConfig("")
}

// Return provider string
func providerString(config *config.Config, alias ...string) string {

	aliasStr := ""
	if len(alias) > 0 {
		aliasStr = "alias = \"" + alias[0] + "\""
	}

	if config == nil {
		config = paralusProviderConfig()
	}

	return fmt.Sprintf(`
		provider "paralusctl" {
			version = "1.0"
			profile = "%s"
			rest_endpoint = "%s"
			ops_endpoint = "%s"
			api_key = "%s"
			api_secret = "%s"
			%s
		}

	`, config.Profile, config.RESTEndpoint, config.OPSEndpoint,
		config.APIKey, config.APISecret, aliasStr)
}

func TestProvider(t *testing.T) {
	if err := provider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = provider.Provider()
}

// makes sure necessary PCTL values are set
func testAccConfigPreCheck(t *testing.T) {

	config := paralusProviderConfig()

	if v := config.Profile; v == "" {
		t.Fatal("PCTL_PROFILE env var or config value must be set for acceptance tests")
	}
	if v := config.RESTEndpoint; v == "" {
		t.Fatal("PCTL_REST_ENDPOINT env var or config value must be set for acceptance tests")
	}

	if v := config.OPSEndpoint; v == "" {
		t.Fatal("PCTL_OPS_ENDPOINT env value or config value must be set for acceptance tests")
	}

	if v := config.APIKey; v == "" {
		t.Fatal("PCTL_API_KEY env var or config value must be set for acceptance tests")
	}

	if v := config.APISecret; v == "" {
		t.Fatal("PCTL_API_SECRET env var or config value must be set for acceptance tests")
	}
}
