// Provider acceptance test
package acctest

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/provider"
	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var conf *config.Config

func init() {
	testAccProvider = provider.Provider()
	testAccProviders = map[string]*schema.Provider{
		"paralus": testAccProvider,
	}
	// conf = paralusProviderConfig()

}

// Return test provider config
func paralusProviderConfig() *config.Config {
	configJson := os.Getenv("PCTL_CONFIG_JSON")
	newConfig, err := paralus.NewConfigFromFile(configJson)
	if err != nil {
		panic(err)
	}
	return newConfig
}

// Return provider string
func providerString(conf *config.Config, alias ...string) string {

	aliasStr := ""
	if len(alias) > 0 {
		aliasStr = "alias = \"" + alias[0] + "\""
	}

	if conf == nil {
		conf = paralusProviderConfig()
	}

	return fmt.Sprintf(`
		provider "paralus" {
			version = "1.0"
			pctl_profile = "%s"
			pctl_rest_endpoint = "%s"
			pctl_ops_endpoint = "%s"
			pctl_api_key = "%s"
			pctl_api_secret = "%s"
			pctl_partner = "%s"
			pctl_organization = "%s"
			%s
		}

	`, conf.Profile, conf.RESTEndpoint, conf.OPSEndpoint,
		conf.APIKey, conf.APISecret, conf.Partner, conf.Organization, aliasStr)
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

	if v := conf.Profile; v == "" {
		t.Fatal("PCTL_PROFILE env var or config value must be set for acceptance tests")
	}
	if v := conf.RESTEndpoint; v == "" {
		t.Fatal("PCTL_REST_ENDPOINT env var or config value must be set for acceptance tests")
	}

	if v := conf.OPSEndpoint; v == "" {
		t.Fatal("PCTL_OPS_ENDPOINT env value or config value must be set for acceptance tests")
	}

	if v := conf.APIKey; v == "" {
		t.Fatal("PCTL_API_KEY env var or config value must be set for acceptance tests")
	}

	if v := conf.APISecret; v == "" {
		t.Fatal("PCTL_API_SECRET env var or config value must be set for acceptance tests")
	}
}

// Test invaid API provider endpoint
func TestAccProviderAttr_setInvalidAPISecret(t *testing.T) {

	resource.Test(t, resource.TestCase{
		// PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccProvider_setInvalidSecret(),
				ExpectError: regexp.MustCompile(".*{\"code\":13,\"message\":\"Internal\"}.*"),
			},
		},
	})
}

// Test invaid API provider endpoint
func TestAccProviderAttr_setInvalidAPIKey(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccProvider_setInvalidAPIKey(),
				ExpectError: regexp.MustCompile(".*{\"code\":13,\"message\":\"Internal\"}.*"),
			},
		},
	})
}

// Test invaid API provider endpoint
func TestAccProviderAttr_setInvalidEndpoint(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccProvider_setRestEndpoint("console.paralus.blahblahblah.com"),
				ExpectError: regexp.MustCompile(".*no such host.*"),
			},
		},
	})
}

// set invaid api secret in provider
func testAccProvider_setInvalidSecret() string {

	conf = paralusProviderConfig()
	conf.APISecret = "smackety"

	return fmt.Sprintf(`
%s

data "paralus_project" "default" {
	provider = paralus.custom_api_secret
	name = "default"
  }`, providerString(conf, "custom_api_secret"))
}

// set invaid api key in provider
func testAccProvider_setInvalidAPIKey() string {

	conf = paralusProviderConfig()
	conf.APIKey = "yackity"

	return fmt.Sprintf(`
%s

data "paralus_project" "default" {
	provider = paralus.custom_api_key
	name = "default"
  }`, providerString(conf, "custom_api_key"))
}

// set rest endpoint value in provider
func testAccProvider_setRestEndpoint(endpoint string) string {

	conf = paralusProviderConfig()
	conf.RESTEndpoint = endpoint

	return fmt.Sprintf(`
%s

resource "paralus_cluster" "default" {
	provider = paralus.custom_rest_endpoint
	name     = "tf-cluster-test"
	project = "blah1"
  }`, providerString(conf, "custom_rest_endpoint"))
}

func TestAccProviderEndpoints_setInvalidPartner(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderEndpoints_setPartner("howdy"),
				ExpectError: regexp.MustCompile(".*no such host.*"),
			},
		},
	})
}

// set partner value in provider
func testAccProviderEndpoints_setPartner(partner string) string {

	conf = paralusProviderConfig()
	conf.Partner = partner

	return fmt.Sprintf(`
%s

resource "paralus_cluster" "default" {
	provider = paralus.custom_partner
	name     = "tf-cluster-test"
	project = "default"
  }`, providerString(conf, "custom_partner"))
}
