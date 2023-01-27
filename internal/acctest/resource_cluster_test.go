// Cluster Resource acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/cluster"
)

// Test project and cluster creation
func TestAccParalusResourceProjectCluster_full(t *testing.T) {

	projectRsName := "paralus_project.testproject"
	clusterRsName := "paralus_cluster.testcluster"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_project" "testproject" {
					provider = paralus.valid_resource
					name = "projectresource"
					description = "from acct test"
				}
		
				resource "paralus_cluster" "testcluster" {
					provider = paralus.valid_resource
					name = "clusterresource"
					project = paralus_project.testproject.name
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}
				`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceClusterExists(clusterRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "from acct test"),
					testAccCheckResourceClusterTypeAttribute(clusterRsName, "imported"),
					resource.TestCheckResourceAttr(projectRsName, "description", "from acct test"),
					resource.TestCheckResourceAttr(clusterRsName, "project", "projectresource"),
					resource.TestCheckTypeSetElemAttr(clusterRsName, "bootstrap_files.*", "12"),
				),
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      clusterRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Test unknown project
func TestAccParalusResourceClusterUnknownProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "test"
					project = "blah"
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}`),
				ExpectError: regexp.MustCompile(".*Project .* does not exist.*"),
			},
		},
	})
}

// Test cluster creation into existing project
func TestAccParalusResourceCluster_basic(t *testing.T) {
	clusterRsName := "paralus_cluster.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "test"
					project = "default"
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceClusterExists(clusterRsName),
					testAccCheckResourceClusterTypeAttribute(clusterRsName, "imported"),
					resource.TestCheckResourceAttr(clusterRsName, "project", "default"),
				),
			},
			{
				ResourceName:      clusterRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
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

// testAccCheckResourceClusterExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Cluster instance
func testAccCheckResourceClusterExists(resourceName string) func(s *terraform.State) error {

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

// testAccCheckResourceClusterTypeAttribute verifies project attribute is set correctly by
// Terraform
func testAccCheckResourceClusterTypeAttribute(resourceName string, cluster_type string) func(s *terraform.State) error {

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
