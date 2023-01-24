package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/cluster"
)

// Test cluster not found
func TestAccParalusCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceClusterConfig("blah"),
				ExpectError: regexp.MustCompile(".*cluster not found.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceCluster_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterConfig("ignoreme"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceClusterExists("data.paralus_cluster.test"),
					testAccCheckDataSourceClusterTypeAttribute("data.paralus_cluster.test", "ignoreme"),
					resource.TestCheckResourceAttr("data.paralus_cluster.test", "project", "default"),
				),
			},
		},
	})
}

func testAccDataSourceClusterConfig(clusterName string) string {

	return fmt.Sprintf(`
		data "paralus_cluster" "test" {
			name = "%s"
			project = "default"
		}
	`, clusterName)
}

// testAccCheckClusterExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Cluster instance
func testAccCheckDataSourceClusterExists(resourceName string) func(s *terraform.State) error {

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
func testAccCheckDataSourceClusterTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.Attributes["description"] != description {
			return fmt.Errorf("Invalid description")
		}

		return nil
	}
}
