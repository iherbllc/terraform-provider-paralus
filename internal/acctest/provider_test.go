// Provider acceptance test
package acctest

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/paralus/cli/pkg/config"

	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"paralus": providerserver.NewProtocol6WithError(provider.New()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	conf = paralusProviderConfig()
	testAccConfigPreCheck(t)
}

// var testAccProviders map[string]*provider.Provider
// var testAccProvider *provider.Provider
var conf *config.Config

// func init() {
// 	testAccProvider = provider.New()
// 	testAccProviders = map[string]*schema.Provider{
// 		"paralus": testAccProvider,
// 	}
// 	conf = paralusProviderConfig()

// }

// Return test provider config
func paralusProviderConfig() *config.Config {
	configJson := os.Getenv("CONFIG_JSON")
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

// func TestProvider(t *testing.T) {
// 	if err := provider.New().InternalValidate(); err != nil {
// 		t.Fatalf("err: %s", err)
// 	}
// }

// func TestProvider_impl(t *testing.T) {
// 	var _ *schema.Provider = provider.New()
// }

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

// // Test invalid provider API secret
// func TestAccProviderAttr_setInvalidAPISecret(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccProvider_setInvalidSecret(),
// 				ExpectError: regexp.MustCompile(".*{\"code\":13,\"message\":\"Internal\"}.*"),
// 			},
// 		},
// 	})
// }

// // Test invalid provider API key
// func TestAccProviderAttr_setInvalidAPIKey(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccProvider_setInvalidAPIKey(),
// 				ExpectError: regexp.MustCompile(".*{\"code\":13,\"message\":\"Internal\"}.*"),
// 			},
// 		},
// 	})
// }

// // Test missing provider API key
// func TestAccProviderAttr_setMissingAPIKey(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccProvider_setMissingAPIKey(),
// 				ExpectError: regexp.MustCompile(".*api key not defined.*"),
// 			},
// 		},
// 	})
// }

// // Test invalid  provider API endpoint
// func TestAccProviderAttr_setInvalidEndpoint(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccProvider_setRestEndpoint("console.paralus.blahblahblah.com"),
// 				ExpectError: regexp.MustCompile(".*(no such host|resource does not exist).*"),
// 			},
// 		},
// 	})
// }

// // Test missing provider API endpoint
// func TestAccProviderAttr_setEmptyEndpoint(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccProvider_setRestEndpoint(""),
// 				ExpectError: regexp.MustCompile(".*(rest endpoint not defined|resource does not exist).*"),
// 			},
// 		},
// 	})
// }

// // Test overriding config json with a bad path
// func TestAccProviderCreds_BadConfigJsonPath(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: `provider "paralus" {
// 					alias = "bad_config_json"
// 					pctl_config_json = "mybad.json"
// 				}

// 				data "paralus_project" "default" {
// 					provider = paralus.bad_config_json
// 					name     = "blah2"
// 				  }
// 				`,
// 				ExpectError: regexp.MustCompile(".*error parsing config_json file.*"),
// 			},
// 		},
// 	})
// }

// // set invalid api secret in provider
// func testAccProvider_setInvalidSecret() string {

// 	conf = paralusProviderConfig()
// 	conf.APISecret = "smackety"

// 	return fmt.Sprintf(`
// %s

// data "paralus_project" "custom_api_secret" {
// 	provider = paralus.custom_api_secret
// 	name = "acctest-donotdelete"
//   }`, providerString(conf, "custom_api_secret"))
// }

// // set invalid api key in provider
// func testAccProvider_setInvalidAPIKey() string {

// 	conf = paralusProviderConfig()
// 	conf.APIKey = "yackity"

// 	return fmt.Sprintf(`
// %s

// data "paralus_project" "custom_api_key" {
// 	provider = paralus.custom_api_key
// 	name = "acctest-donotdelete"
//   }`, providerString(conf, "custom_api_key"))
// }

// // set missing api key in provider
// func testAccProvider_setMissingAPIKey() string {

// 	conf = paralusProviderConfig()
// 	conf.APIKey = ""

// 	return fmt.Sprintf(`
// %s

// data "paralus_project" "missing_api_key" {
// 	provider = paralus.missing_api_key
// 	name = "acctest-donotdelete"
//   }`, providerString(conf, "missing_api_key"))
// }

// // set rest endpoint value in provider
// func testAccProvider_setRestEndpoint(endpoint string) string {

// 	conf = paralusProviderConfig()
// 	conf.RESTEndpoint = endpoint

// 	return fmt.Sprintf(`
// %s

// resource "paralus_cluster" "custom_rest_endpoint" {
// 	provider = paralus.custom_rest_endpoint
// 	name     = "tf-cluster-test"
// 	project = "blah1"
// 	cluster_type = "imported"
//   }`, providerString(conf, "custom_rest_endpoint"))
// }

// Set a valid provider
func testAccProviderValidResource(resources string) string {
	conf = paralusProviderConfig()
	return fmt.Sprintf(`
%s

%s`, providerString(conf, "valid_resource"), resources)
}
