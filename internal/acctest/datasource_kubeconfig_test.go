// Cluster DataSource bootstrap file acceptance test
package acctest

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
)

// Test cluster not found
func TestAccParalusKubeconfigUserNotFound_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceKubeconfigNoClusterConfig("blah"),
				ExpectError: regexp.MustCompile(".*error locating user info.*"),
			},
		},
	})
}

func TestAccParalusKubeconfigConfigNotFound_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceKubeconfigNoClusterConfig("acctest-user@example.com"),
				ExpectError: regexp.MustCompile(".*error locating kubeconfig for user.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceKubeconfig_basic(t *testing.T) {
	dsResourceName := "data.paralus_kubeconfig.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubeconfigNoClusterConfig("acctest2-user@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHasKubeConfig(dsResourceName),
					testAccCheckResourceAttributeSet(dsResourceName, "client_certificate_data"),
					testAccCheckResourceAttributeSet(dsResourceName, "client_key_data"),
					testAccCheckDataSourceKubeConfigAttributeNotNil(dsResourceName),
					resource.TestCheckTypeSetElemAttr(dsResourceName, "cluster_info.*", "1"),
				),
			},
		},
	})
}

func testAccDataSourceKubeconfigNoClusterConfig(name string) string {

	return fmt.Sprintf(`
		data "paralus_kubeconfig" "test" {
			name = "%s"
		}`, name)
}

// Standard acceptance test with cluster name
func TestAccParalusDataSourceKubeconfigWithCluster_basic(t *testing.T) {
	dsResourceName := "data.paralus_kubeconfig.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubeconfigWithClusterConfig("acctest2-user@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHasKubeConfig(dsResourceName),
					testAccCheckResourceAttributeSet(dsResourceName, "client_certificate_data"),
					testAccCheckResourceAttributeSet(dsResourceName, "client_key_data"),
					testAccCheckDataSourceKubeConfigAttributeNotNil(dsResourceName),
					resource.TestCheckResourceAttr(dsResourceName, "cluster", "minikube"),
					resource.TestCheckTypeSetElemAttr(dsResourceName, "cluster_info.*", "1"),
				),
			},
		},
	})
}

func testAccDataSourceKubeconfigWithClusterConfig(name string) string {

	return fmt.Sprintf(`
		data "paralus_kubeconfig" "test" {
			name = "%s"
			cluster = "minikube"
		}`, name)
}

// tests whether a kubeconfig file exists
func testAccCheckHasKubeConfig(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Name is not set")
		}

		cluster := rs.Primary.Attributes["cluster"]
		name := rs.Primary.Attributes["name"]
		userInfo, err := utils.GetUserByName(context.Background(), name, nil)

		if err != nil {
			return err
		}

		_, err = utils.GetKubeConfig(context.Background(), userInfo.Metadata.Id, cluster, "", nil)

		if err != nil {
			return err
		}
		return nil
	}
}

// Verifies project attribute is set correctly by Terraform
func testAccCheckDataSourceKubeConfigAttributeNotNil(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		i, err := strconv.Atoi(rs.Primary.Attributes["cluster_info.#"])
		if err != nil || i <= 0 {
			return fmt.Errorf("requested cluster info not found")
		}

		return nil
	}
}
