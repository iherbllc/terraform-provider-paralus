// Package DataSource group acceptance test
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

// Test group not found
func TestAccParalusNoGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceGroupConfig("blah"),
				ExpectError: regexp.MustCompile(".*resource does not exist.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceGroup_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

// Uses the paralus API through PCTL to retrieve group info
// and store it as a PCTL Group instance
func testAccCheckDataSourceGroupExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("group id is not set")
		}

		groupStr := rs.Primary.Attributes["name"]

		_, err := utils.GetGroupByName(context.Background(), groupStr, nil)

		if err != nil {
			return err
		}
		return nil
	}
}

// Verifies group attribute is set correctly by Terraform
func testAccCheckDataSourceGroupTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.Attributes["description"] != description {
			return fmt.Errorf("invalid description")
		}

		return nil
	}
}
