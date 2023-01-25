// Project Resource acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/project"
)

// Test fail create project if organization name not same as UI configuration
func TestAccParalusResourceProjectBadOrg_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectResourceConfigBadOrg(),
				ExpectError: regexp.MustCompile(".*not authorized to perform action.*"),
			},
		},
	})
}

func testAccProjectResourceConfigBadOrg() string {

	conf = paralusProviderConfig()
	conf.Organization = "blah"

	providerConfig := providerString(conf, "project_badorg_test")
	return fmt.Sprintf(`
		%s

		resource "paralus_project" "badorg_test" {
			provider = paralus.project_badorg_test
			name = "badorg_project"
		}
	`, providerConfig)
}

// General Paralus project resource creation
func TestAccParalusResourceProject_basic(t *testing.T) {

	projectRsName := "paralus_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "paralus_project" "test" {
					name = "test"
					description = "test project"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "test project"),
				),
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccCheckProjectResourceDestroy verifies the cluster has been destroyed
func testAccCheckProjectResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_project" {
				continue
			}

			projectStr := rs.Primary.Attributes["name"]

			_, err := project.GetProjectByName(projectStr)

			if err == nil {
				project.DeleteProject(projectStr)
			}
		}

		return nil
	}
}

// testAccCheckProjectExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Project instance
func testAccCheckResourceProjectExists(resourceName string) func(s *terraform.State) error {

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

// testAccCheckProjectTypeAttribute verifies project attribute is set correctly by
// Terraform
func testAccCheckResourceProjectTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

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
