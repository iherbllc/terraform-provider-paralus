package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/project"
)

func TestAccParalusProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists("paralus_project.test"),
					testAccCheckProjectTypeAttribute("paralus_project.test", "test project"),
					resource.TestCheckResourceAttr("paralus_project.test", "description", "test project"),
				),
			},
		},
	})
}

func testAccProjectConfig() string {

	conf = paralusProviderConfig()

	providerConfig := providerString(conf, "project_test")
	return fmt.Sprintf(`
		%s

		resource "paralus_project" "test" {
			provider = "paralusctl.project_test"
			name = "test"
			description = "test project"
		}
	`, providerConfig)
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
func testAccCheckProjectExists(resourceName string) func(s *terraform.State) error {

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
func testAccCheckProjectTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

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
