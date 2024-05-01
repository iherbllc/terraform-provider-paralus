// Cluster DataSource cluster acceptance test
package acctest

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
)

// Test cluster not found
func TestAccParalusClusterNotFound_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceClusterConfig("blah"),
				ExpectError: regexp.MustCompile(".*error locating cluster.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceCluster_basic(t *testing.T) {
	dsResourceName := "data.paralus_cluster.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterConfig("man-acctest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceClusterExists(dsResourceName),
					testAccCheckDataSourceClusterTypeAttribute(dsResourceName, "acctest-donotdelete"),
					testAccCheckResourceAttributeSet(dsResourceName, "relays"),
					testAccCheckResourceAttributeSet(dsResourceName, "uuid"),
					resource.TestCheckResourceAttr(dsResourceName, "project", "acctest-donotdelete"),
					resource.TestCheckTypeSetElemAttr(dsResourceName, "bootstrap_files.*", "12"),
				),
			},
		},
	})
}

func testAccDataSourceClusterConfig(clusterName string) string {

	return fmt.Sprintf(`
		data "paralus_cluster" "test" {
			name = "%s"
			project = "acctest-donotdelete"
		}
	`, clusterName)
}

// Uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Cluster instance
func testAccCheckDataSourceClusterExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("cluster id is not set")
		}

		project := rs.Primary.Attributes["project"]
		clusterName := rs.Primary.Attributes["name"]

		_, err := utils.GetCluster(context.Background(), clusterName, project, nil)

		if err != nil {
			return err
		}
		return nil
	}
}

// Verifies project attribute is set correctly by Terraform
func testAccCheckDataSourceClusterTypeAttribute(resourceName string, project string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.Attributes["project"] != project {
			return fmt.Errorf("invalid project")
		}

		return nil
	}
}

// Full acceptance test with datasource for both cluster and project
func TestAccParalusDataSourceCluster_Full(t *testing.T) {
	dsResourceClusterName := "data.paralus_cluster.test"
	dsResourceProjectName := "data.paralus_project.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterFullConfig("man-acctest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceClusterExists(dsResourceClusterName),
					testAccCheckDataSourceClusterTypeAttribute(dsResourceClusterName, "acctest-donotdelete"),
					testAccCheckResourceAttributeSet(dsResourceClusterName, "relays"),
					testAccCheckResourceAttributeSet(dsResourceClusterName, "uuid"),
					resource.TestCheckResourceAttr(dsResourceClusterName, "project", "acctest-donotdelete"),
					resource.TestCheckTypeSetElemAttr(dsResourceClusterName, "bootstrap_files.*", "12"),
					testAccCheckDataSourceProjectExists(dsResourceProjectName),
					testAccCheckDataSourceProjectTypeAttribute(dsResourceProjectName, "Project used for acceptance testing"),
					resource.TestCheckResourceAttr(dsResourceProjectName, "description", "Project used for acceptance testing"),
				),
			},
		},
	})
}

func testAccDataSourceClusterFullConfig(clusterName string) string {

	return fmt.Sprintf(`
		data "paralus_project" "test" {
			name = "acctest-donotdelete"
		}
		data "paralus_cluster" "test" {
			name = "%s"
			project = data.paralus_project.test.name
		}
	`, clusterName)
}
