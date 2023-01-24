package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/cluster"
)

func TestAccParalusCluster_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("paralus_cluster.test"),
					testAccCheckClusterTypeAttribute("paralus_cluster.test", "imported"),
					resource.TestCheckResourceAttr("paralus_cluster.test", "project", "default"),
				),
			},
		},
	})
}

func testAccClusterConfig() string {

	conf = paralusProviderConfig()

	providerConfig := providerString(conf, "cluster_test")
	return fmt.Sprintf(`
		%s

		resource "paralus_cluster" "test" {
			provider = "paralusctl.cluster_test"
			name = "test"
			description = "test cluster"
			project = "default"
			cluster_type = "imported"
			params {
				provision_type = "IMPORT"
				provision_environment = "CLOUD"
				kubernetes_provider = "EKS"
				state = "PROVISION"
			}
		}
	`, providerConfig)
}

// testAccCheckClusterResourceDestroy verifies the cluster has been destroyed
func testAccCheckClusterResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_cluster" {
				continue
			}

			project := rs.Primary.Attributes["project"]
			clusterName := rs.Primary.Attributes["name"]

			_, err := cluster.GetCluster(clusterName, project)

			if err == nil {
				cluster.DeleteCluster(clusterName, project)
			}
		}

		return nil
	}
}

// testAccCheckClusterExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Cluster instance
func testAccCheckClusterExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Cluster ID is not set")
		}

		project := rs.Primary.Attributes["project"]
		clusterName := rs.Primary.Attributes["name"]

		_, err := cluster.GetCluster(clusterName, project)

		if err == nil {
			return err
		}
		return nil
	}
}

// testAccCheckClusterTypeAttribute verifies project attribute is set correctly by
// Terraform
func testAccCheckClusterTypeAttribute(resourceName string, cluster_type string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.Attributes["cluster_type"] != cluster_type {
			return fmt.Errorf("Invalid cluster type")
		}

		return nil
	}
}
