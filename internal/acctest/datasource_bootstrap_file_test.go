// Cluster DataSource bootstrap file acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/cluster"
)

// Test cluster not found
func TestAccParalusBootstrapNotFound_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceBootstrapConfig("blah"),
				ExpectError: regexp.MustCompile(".*Error locating cluster.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceBootstrap_basic(t *testing.T) {
	dsResourceName := "data.paralus_bootstrap_file.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceBootstrapConfig("man-acctest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHasBootstrap(dsResourceName),
					testAccCheckResourceAttributeSet(dsResourceName, "relays"),
					testAccCheckResourceAttributeSet(dsResourceName, "uuid"),
					testAccCheckDataSourceBootstrapAttributeNotNil(dsResourceName),
					resource.TestCheckTypeSetElemAttr(dsResourceName, "bootstrap_files.*", "12"),
				),
			},
		},
	})
}

func testAccDataSourceBootstrapConfig(clusterName string) string {

	return fmt.Sprintf(`
		data "paralus_bootstrap_file" "test" {
			name = "%s"
			project = "acctest-donotdelete"
		}
	`, clusterName)
}

// tests whether a bootstrap file exists
func testAccCheckHasBootstrap(resourceName string) func(s *terraform.State) error {

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

		_, err := cluster.GetBootstrapFile(clusterName, project)

		if err != nil {
			return err
		}
		return nil
	}
}

// Verifies project attribute is set correctly by Terraform
func testAccCheckDataSourceBootstrapAttributeNotNil(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.Attributes["bootstrap_files_combined"] == "" {
			return fmt.Errorf("No bootstrap provided")
		}
		i, err := strconv.Atoi(rs.Primary.Attributes["bootstrap_files.#"])
		if err != nil || i <= 0 {
			return fmt.Errorf("No bootstrap files provided")
		}

		return nil
	}
}
