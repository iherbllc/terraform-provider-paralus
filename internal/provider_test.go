package acctest

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/iherbllc/terraform-provider-paralus/internal/paralus"
	"github.com/iherbllc/terraform-provider-paralus/internal/provider"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/joho/godotenv"
	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

// credential ENV vars
var credsEnvVars = []string{
	"PCTL_API_KEY",
	"PCTL_API_SECRET",
}

// endpoint ENV vars
var endpointEnvVars = []string{
	"PCTL_REST_ENDPOINT",
	"PCTL_OPS_ENDPOINT",
}

var configEnvVars = []string{
	"PCTL_PROFILE",
}

func init() {
	testAccProvider = provider.Provider()
	testAccProviders = map[string]*schema.Provider{
		"paralus": testAccProvider,
	}
}

// Return test provider config
func paralusProviderConfig() *config.Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		return paralus.NewConfig()
	}
	return nil
}

func TestProvider(t *testing.T) {
	if err := provider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = provider.Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := utils.MultiEnvSearch(credsEnvVars); v == "" {
		t.Fatalf("One of %s must be set for acceptance tests", strings.Join(credsEnvVars, ", "))
	}

	if v := utils.MultiEnvSearch(configEnvVars); v == "" {
		t.Fatalf("One of %s must be set for acceptance tests", strings.Join(configEnvVars, ", "))
	}

	if v := utils.MultiEnvSearch(endpointEnvVars); v == "" {
		t.Fatalf("One of %s must be set for acceptance tests", strings.Join(endpointEnvVars, ", "))
	}
}

// testAccPreCheck ensures at least one of the config env variables is set.
func getTestConfigsFromEnv() string {
	return utils.MultiEnvSearch(configEnvVars)
}

// testAccPreCheck ensures at least one of the credentials env variables is set.
func getTestCredsFromEnv() string {
	return utils.MultiEnvSearch(credsEnvVars)
}

// testAccPreCheck ensures at least one of the endpoint env variables is set.
func getTestEndpointsFromEnv() string {
	return utils.MultiEnvSearch(endpointEnvVars)
}

func TestAccProviderEndpoints_setInvalidRestEndpoints(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderEndpoints_setRestEndpoint("https://www.example.com", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("got HTTP response code 404 with body"),
			},
		},
	})
}

func testAccProviderEndpoints_setRestEndpoint(endpoint, name string) string {
	return fmt.Sprintf(`
provider "paralus" {
  alias                   = "custom_rest_endpoint"
  rest_endpoint = "%s"
}

resource "cluster" "default" {
	provider = paralus.custom_rest_endpoint
	name     = "tf-cluster-%s"
  }`, endpoint, name)
}

func testAccProviderEndpoints_setOpsEndpoint(endpoint, name string) string {
	return fmt.Sprintf(`
provider "google" {
  alias                   = "custom_ops_endpoint"
  ops_endpoint = "%s"
}

resource "cluster" "default" {
	provider = paralus.custom_ops_endpoint
	name     = "tf-cluster-%s"
  }`, endpoint, name)
}
