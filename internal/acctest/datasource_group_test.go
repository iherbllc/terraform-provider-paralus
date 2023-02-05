// Package DataSource group acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/group"
)

// Test group not found
func TestAccParalusNoGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceGroupConfig("blah"),
				ExpectError: regexp.MustCompile(".*no rows in result set.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceGroup_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroupConfig("All Local Users"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceGroupExists("data.paralus_group.default"),
					testAccCheckDataSourceGroupTypeAttribute("data.paralus_group.default", "Default group for all local users"),
					resource.TestCheckResourceAttr("data.paralus_group.default", "description", "Default group for all local users"),
				),
			},
		},
	})
}

func testAccDataSourceGroupConfig(groupName string) string {

	return fmt.Sprintf(`
		data "paralus_group" "default" {
			name = "%s"
		}
	`, groupName)
}

// testAccCheckDataSourceGroupExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Group instance
func testAccCheckDataSourceGroupExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Group ID is not set")
		}

		groupStr := rs.Primary.Attributes["name"]

		_, err := group.GetGroupByName(groupStr)

		if err != nil {
			return err
		}
		return nil
	}
}

// testAccCheckDataSourceGroupTypeAttribute verifies group attribute is set correctly by
// Terraform
func testAccCheckDataSourceGroupTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

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