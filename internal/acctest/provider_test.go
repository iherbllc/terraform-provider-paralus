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
		"paralusctl": testAccProvider,
	}
}

// Return test provider config
func paralusProviderConfig() *config.Config {
	// utils.LoadEnv()
	newConfig, err := paralus.NewConfigFromFile(os.Getenv("CONFIG_JSON"))
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

	return fmt.Sprintf(`
		provider "paralusctl" {
			version = "1.0"
			profile = "%s"
			rest_endpoint = "%s"
			ops_endpoint = "%s"
			api_key = "%s"
			api_secret = "%s"
			partner = "%s"
			organization = "%s"
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

func TestAccProviderEndpoints_setInvalidRestEndpoints(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderEndpoints_setRestEndpoint("console.paralus.blahblahblah.com"),
				ExpectError: regexp.MustCompile(".*no such host.*"),
			},
		},
	})
}

// set rest endpoint value in provider
func testAccProviderEndpoints_setRestEndpoint(endpoint string) string {

	// var conf *config.Config
	if testAccProvider.Meta() == nil {
		conf = paralusProviderConfig()
	} else {
		conf = testAccProvider.Meta().(*config.Config)
	}

	conf.RESTEndpoint = endpoint

	return fmt.Sprintf(`
%s

resource "paralus_cluster" "default" {
	provider = paralusctl.custom_rest_endpoint
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

	// var conf *config.Config
	if testAccProvider.Meta() == nil {
		conf = paralusProviderConfig()
	} else {
		conf = testAccProvider.Meta().(*config.Config)
	}

	conf.Partner = partner

	return fmt.Sprintf(`
%s

resource "paralus_cluster" "default" {
	provider = paralusctl.custom_partner
	name     = "tf-cluster-test"
	project = "default"
  }`, providerString(conf, "custom_partner"))
}
