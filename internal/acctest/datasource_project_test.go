package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/project"
)

// Test project not found
func TestAccParalusNoProject_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceProjectConfig("blah"),
				ExpectError: regexp.MustCompile(".*no rows in result set.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceProjectConfig("default"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceProjectExists("data.paralus_project.default"),
					testAccCheckDataSourceProjectTypeAttribute("data.paralus_project.default", "Default project"),
					resource.TestCheckResourceAttr("data.paralus_project.default", "description", "Default project"),
				),
			},
		},
	})
}

func testAccDataSourceProjectConfig(projectName string) string {

	conf = paralusProviderConfig()

	providerConfig := providerString(conf, "project_ds_test")
	return fmt.Sprintf(`
		%s

		data "paralus_project" "default" {
			provider = "paralusctl.project_ds_test"
			name = "%s"
		}
	`, providerConfig, projectName)
}

// testAccCheckDataSourceProjectExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Project instance
func testAccCheckDataSourceProjectExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Project ID is not set")
		}

		projectStr := rs.Primary.Attributes["name"]

		_, err := project.GetProjectByName(projectStr)

		if err == nil {
			return err
		}
		return nil
	}
}

// testAccCheckDataSourceProjectTypeAttribute verifies project attribute is set correctly by
// Terraform
func testAccCheckDataSourceProjectTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

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
