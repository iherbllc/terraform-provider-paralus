// Project Resource acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/paralus/cli/pkg/project"
)

// Test missing project name
func TestAccParalusResourceMissingProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectResourceConfigMissingProject(),
				ExpectError: regexp.MustCompile(".*argument \"name\" is required.*"),
			},
		},
	})
}

func testAccProjectResourceConfigMissingProject() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "project_missing_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_project" "missingname_test" {
			provider = paralus.project_missing_name
		}
	`, providerConfig)
}

// Test empty project name
func TestAccParalusResourceEmptyProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectResourceConfigEmptyProject(),
				ExpectError: regexp.MustCompile(".*name cannot be empty.*"),
			},
		},
	})
}

func testAccProjectResourceConfigEmptyProject() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "project_empty_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_project" "emptyname_test" {
			provider = paralus.project_empty_name
			name = ""
		}
	`, providerConfig)
}

// Test fail create project if organization name not same as UI configuration
func TestAccParalusResourceProjectBadOrg_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectResourceConfigBadOrg(),
				ExpectError: regexp.MustCompile(".*Failed to create project.*"),
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
				Config: testAccProviderValidResource(`
				resource "paralus_project" "test" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
				}`),
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

		if err != nil {
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

// Test creating project and adding in group
func TestAccParalusResourceProject_AddToGroup(t *testing.T) {
	groupRsName := "paralus_group.test"
	projectRsName := "paralus_project.add_to_group"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				// we will have a non-empty plan because the project access removal will affect the group role as well
				ExpectNonEmptyPlan: true,
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test"
					description = "test group"
				}

				resource "paralus_project" "add_to_group" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
					project_roles {
						project = "test"
						role = "PROJECT_READ_ONLY"
						group = paralus_group.test.name
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName),
					testAccCheckResourceGroupTypeAttribute(groupRsName, "test group"),
					resource.TestCheckResourceAttr(groupRsName, "description", "test group"),
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "test project"),
					testAccCheckResourceProjectProjectRoleMap(projectRsName, map[string]string{"role": "PROJECT_READ_ONLY"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{
						"role":    "PROJECT_READ_ONLY",
						"group":   "test",
						"project": "test",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
				),
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      groupRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Test creating project and adding two groups
func TestAccParalusResourceProject_Add2Groups(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test1" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test group1"
				}

				resource "paralus_group" "test2" {
					provider = paralus.valid_resource
					name = "test2"
					description = "test group2"
				}

				resource "paralus_project" "add_to_group" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
					project_roles {
						project = "test"
						role = "PROJECT_READ_ONLY"
						group = paralus_group.test1.name
					}
					project_roles {
						project = "test"
						role = "PROJECT_READ_ONLY"
						group = paralus_group.test2.name
					}
				}`),
				ExpectError: regexp.MustCompile("permissions must be distinct between project_roles blocks.*"),
			},
		},
	})
}

// testAccCheckResourceProjectProjectRoleMap verifies project role list for project
func testAccCheckResourceProjectProjectRoleMap(resourceName string, projectRoles map[string]string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		projectStr := rs.Primary.Attributes["name"]

		projectStruct, err := project.GetProjectByName(projectStr)

		if err != nil {
			return err
		}

		return utils.ValidateProjectNamespaceRolesSet(projectStruct.Spec.ProjectNamespaceRoles, projectRoles)
	}
}
