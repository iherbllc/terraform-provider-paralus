// Package DataSource project acceptance test
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

// Test missing project name
func TestAccParalusDataSourceMissingProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// CheckDestroy: testAccCheckProjectDataSourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectDataSourceConfigMissingProject(),
				ExpectError: regexp.MustCompile(".*argument \"name\" is required.*"),
			},
		},
	})
}

func testAccProjectDataSourceConfigMissingProject() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "project_missing_name")
	return fmt.Sprintf(`
		%s

		data "paralus_project" "missingname_test" {
			provider = paralus.project_missing_name
		}
	`, providerConfig)
}

// Test empty project name
func TestAccParalusDataSourceEmptyProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// CheckDestroy: testAccCheckProjectDataSourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectDataSourceConfigEmptyProject(),
				ExpectError: regexp.MustCompile(".*expected not empty string.*"),
			},
		},
	})
}

func testAccProjectDataSourceConfigEmptyProject() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "project_empty_name")
	return fmt.Sprintf(`
		%s

		data "paralus_project" "emptyname_test" {
			provider = paralus.project_empty_name
			name = ""
		}
	`, providerConfig)
}

// Test project not found
func TestAccParalusNoProject_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceProjectConfig("blah"),
				ExpectError: regexp.MustCompile(".*resource does not exist.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceProjectConfig("acctest-donotdelete"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceProjectExists("data.paralus_project.default"),
					testAccCheckDataSourceProjectTypeAttribute("data.paralus_project.default", "Project used for acceptance testing"),
					resource.TestCheckResourceAttr("data.paralus_project.default", "description", "Project used for acceptance testing"),
				),
			},
		},
	})
}

func testAccDataSourceProjectConfig(projectName string) string {

	return fmt.Sprintf(`
		data "paralus_project" "default" {
			name = "%s"
		}
	`, projectName)
}

// Uses the paralus API through PCTL to retrieve project info
// and store it as a PCTL Project instance
func testAccCheckDataSourceProjectExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project id is not set")
		}

		projectStr := rs.Primary.Attributes["name"]

		_, err := utils.GetProjectByName(context.Background(), projectStr, nil)

		if err != nil {
			return err
		}
		return nil
	}
}

// Verifies project attribute is set correctly by Terraform
func testAccCheckDataSourceProjectTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

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
