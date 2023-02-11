// Group Resource acceptance test
package acctest

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
)

// Test missing group name
func TestAccParalusResourceMissingGroup_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupResourceConfigMissingGroup(),
				ExpectError: regexp.MustCompile(".*argument \"name\" is required.*"),
			},
		},
	})
}

func testAccGroupResourceConfigMissingGroup() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "group_missing_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_group" "missingname_test" {
			provider = paralus.group_missing_name
		}
	`, providerConfig)
}

// Test empty group name
func TestAccParalusResourceEmptyGroup_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupResourceConfigEmptyGroup(),
				ExpectError: regexp.MustCompile(".*expected not empty string.*"),
			},
		},
	})
}

func testAccGroupResourceConfigEmptyGroup() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "group_empty_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_group" "emptyname_test" {
			provider = paralus.group_empty_name
			name = ""
		}
	`, providerConfig)
}

// Test fail create group if organization name not same as UI configuration
func TestAccParalusResourceGroupBadOrg_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccConfigPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupResourceConfigBadOrg(),
				ExpectError: regexp.MustCompile(".*not authorized to perform action.*"),
			},
		},
	})
}

func testAccGroupResourceConfigBadOrg() string {

	conf = paralusProviderConfig()
	conf.Organization = "blah"

	providerConfig := providerString(conf, "group_badorg_test")
	return fmt.Sprintf(`
		%s

		resource "paralus_group" "badorg_test" {
			provider = paralus.group_badorg_test
			name = "badorg_group"
		}
	`, providerConfig)
}

// General Paralus group resource creation
func TestAccParalusResourceGroup_basic(t *testing.T) {

	groupRsName := "paralus_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test"
					description = "test group"
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName),
					testAccCheckResourceGroupTypeAttribute(groupRsName, "test group"),
					resource.TestCheckResourceAttr(groupRsName, "description", "test group"),
				),
			},
			{
				ResourceName:      groupRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Verifies the cluster has been destroyed
func testAccCheckGroupResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_group" {
				continue
			}

			groupStr := rs.Primary.Attributes["name"]

			_, err := utils.GetGroupByName(groupStr)

			if err == nil {
				utils.DeleteGroup(groupStr)
			}
		}

		return nil
	}
}

// Uses the paralus API through PCTL to retrieve group info
// and store it as a PCTL Group instance
func testAccCheckResourceGroupExists(resourceName string) func(s *terraform.State) error {

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

		group, err := utils.GetGroupByName(groupStr)

		if err != nil {
			return err
		}
		log.Printf("group info %s", group)
		return nil
	}
}

// Verifies group attribute is set correctly by Terraform
func testAccCheckResourceGroupTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.Attributes["description"] != description {
			return fmt.Errorf("Invalid description")
		}

		return nil
	}
}

// Paralus group creation for one project
func TestAccParalusResourceGroup_Project(t *testing.T) {

	groupRsName := "paralus_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test"
					description = "test group"
					project_roles {
						role = "ADMIN"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName),
					testAccCheckResourceGroupTypeAttribute(groupRsName, "test group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName, map[string]string{
						"role":  "ADMIN",
						"group": "test",
					}),
					resource.TestCheckResourceAttr(groupRsName, "description", "test group"),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName, "project_roles.*", map[string]string{"role": "ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName, "project_roles.*", map[string]string{"group": "test"}),
				),
			},
			{
				ResourceName:      groupRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Verifies user is in list from API
func testAccCheckResourceGroupProjectRoleMap(resourceName string, projectRoles map[string]string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		groupStr := rs.Primary.Attributes["name"]

		group, err := utils.GetGroupByName(groupStr)

		if err != nil {
			return err
		}

		return utils.ValidateProjectNamespaceRolesSet(group.Spec.ProjectNamespaceRoles, projectRoles)
	}
}

// Multiple group creation adding to previously created project
func TestAccParalusResourceGroups_AddToProject(t *testing.T) {
	projectRsName := "paralus_project.temp"
	groupRsName1 := "paralus_group.test1"
	groupRsName2 := "paralus_group.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: true,
				Config: testAccProviderValidResource(`
				resource "paralus_project" "temp" {
					name = "tempproject"
					description = "A temporary project"
				}
				resource "paralus_group" "test1" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					project_roles {
						project = paralus_project.temp.name
						role = "PROJECT_ADMIN"
					}
				}
				resource "paralus_group" "test2" {
					provider = paralus.valid_resource
					name = "test2"
					description = "test 2 group"
					project_roles {
						project = paralus_project.temp.name
						role = "PROJECT_ADMIN"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupTypeAttribute(groupRsName1, "test 1 group"),
					resource.TestCheckResourceAttr(groupRsName1, "description", "test 1 group"),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName1, "project_roles.*", map[string]string{"project": "tempproject"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName1, "project_roles.*", map[string]string{"role": "PROJECT_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName1, "project_roles.*", map[string]string{"group": "test1"}),

					testAccCheckResourceGroupExists(groupRsName2),
					testAccCheckResourceGroupTypeAttribute(groupRsName2, "test 2 group"),
					resource.TestCheckResourceAttr(groupRsName2, "description", "test 2 group"),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName2, "project_roles.*", map[string]string{"project": "tempproject"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName2, "project_roles.*", map[string]string{"role": "PROJECT_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName2, "project_roles.*", map[string]string{"group": "test2"}),

					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "A temporary project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "A temporary project"),
					testAccCheckResourceProjectProjectRoleMap(projectRsName, map[string]string{
						"role":    "PROJECT_ADMIN",
						"group":   "test1",
						"project": "tempproject",
					}),
					testAccCheckResourceProjectProjectRoleMap(projectRsName, map[string]string{
						"role":    "PROJECT_ADMIN",
						"group":   "test2",
						"project": "tempproject",
					}),
				),
			},
			{
				ResourceName:      groupRsName1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      groupRsName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Create a group with a user
func TestAccParalusResourceGroup_AddUser(t *testing.T) {
	groupRsName1 := "paralus_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					users = ["acctest-user@example.com"]
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupTypeAttribute(groupRsName1, "test 1 group"),
					testAccCheckResourceGroupCheckUserList(groupRsName1, "acctest-user@example.com"),
					resource.TestCheckResourceAttr(groupRsName1, "description", "test 1 group"),
					resource.TestCheckResourceAttr(groupRsName1, "users.0", "acctest-user@example.com"),
				),
			},
			{
				ResourceName:      groupRsName1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Test adding a non-existing user to a group
func TestAccParalusResourceGroup_AddNonExistingUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					users = ["nobody@here.com"]
				}`),
				ExpectError: regexp.MustCompile(".*does not exist.*"),
			},
		},
	})
}

// Test adding a non-existing project to a group
func TestAccParalusResourceGroup_AddNonExistingProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					project_roles {
						role = "PROJECT_ADMIN"
						project = "i dont exist"
					}
				}`),
				ExpectError: regexp.MustCompile(".*does not exist.*"),
			},
		},
	})
}

// Test requesting a project role witihout specifying a project
func TestAccParalusResourceProject_NoProjectSpecified(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					project_roles {
						role = "PROJECT_ADMIN"
					}
				}`),
				ExpectError: regexp.MustCompile(".*project must be specified.*"),
			},
		},
	})
}

// Verifies user is in list from API
func testAccCheckResourceGroupCheckUserList(resourceName string, user string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		groupStr := rs.Primary.Attributes["name"]

		group, err := utils.GetGroupByName(groupStr)

		if err != nil {
			return err
		}

		if len(group.Spec.Users) <= 0 {
			return fmt.Errorf("user list is empty")
		}

		if group.Spec.Users[0] != user {
			return fmt.Errorf("user list %s is missing %s", group.Spec.Users, user)
		}

		return nil
	}
}
