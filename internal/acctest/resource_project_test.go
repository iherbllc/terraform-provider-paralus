// Project Resource acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
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
				ExpectError: regexp.MustCompile(".*expected not empty string.*"),
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
				ExpectError: regexp.MustCompile(".*failed to create project.*"),
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

// Verifies the cluster has been destroyed
func testAccCheckProjectResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_project" {
				continue
			}

			projectStr := rs.Primary.Attributes["name"]

			_, err := utils.GetProjectByName(projectStr)

			if err == nil {
				clusters, err := utils.ListAllClusters(projectStr)
				if len(clusters) == 0 || err == utils.ErrResourceNotExists {
					utils.DeleteProject(projectStr)
				}
			}
		}

		return nil
	}
}

// Uses the paralus API through PCTL to retrieve project info
// and store it as a PCTL Project instance
func testAccCheckResourceProjectExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project id is not set")
		}

		projectStr := rs.Primary.Attributes["name"]

		_, err := utils.GetProjectByName(projectStr)

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
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.Attributes["description"] != description {
			return fmt.Errorf("invalid description")
		}

		return nil
	}
}

// Test adding a non-existing user to a project
func TestAccParalusResourceProject_AddNonExistingUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_project" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					user_roles {
						user = "nobody@here.com"
						role = "PROJECT_ADMIN"
					}
				}`),
				ExpectError: regexp.MustCompile(".*does not exist.*"),
			},
		},
	})
}

// Test requesting an empty group name for the project roles
func TestAccParalusResourceProject_GroupNameEmpty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_project" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					project_roles {
						group = ""
						role = "PROJECT_ADMIN"
					}
				}`),
				ExpectError: regexp.MustCompile(".*cannot be empty.*"),
			},
		},
	})
}

// Test adding a non-existing group to a project
func TestAccParalusResourceProject_AddNonExistingGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_project" "test" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
					project_roles {
						group = "does not exist"
						role = "PROJECT_ADMIN"
					}
				}`),
				ExpectError: regexp.MustCompile(".*does not exist.*"),
			},
		},
	})
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

// Test creating a project and adding a namespace role group and project role group
func TestAccParalusResourceProject_Add2GroupsNamespaceAndProjectRoles(t *testing.T) {

	groupRsName1 := "paralus_group.test1"
	groupRsName2 := "paralus_group.test2"
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
				resource "paralus_group" "test1" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
				}

				resource "paralus_group" "test2" {
					provider = paralus.valid_resource
					name = "test2"
					description = "test 2 group"
				}

				resource "paralus_project" "add_to_group" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
					project_roles {
						role = "NAMESPACE_READ_ONLY"
						namespace = "platform"
						group = paralus_group.test1.name
					}
					project_roles {
						role = "PROJECT_ADMIN"
						group = paralus_group.test2.name
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "test project"),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"role": "NAMESPACE_READ_ONLY"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"namespace": "platform"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test1"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"role": "PROJECT_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test2"}),

					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupTypeAttribute(groupRsName1, "test 1 group"),
					resource.TestCheckResourceAttr(groupRsName1, "description", "test 1 group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName1, map[string]string{
						"role":      "NAMESPACE_READ_ONLY",
						"group":     "test1",
						"project":   "test",
						"namespace": "platform",
					}),

					testAccCheckResourceGroupExists(groupRsName2),
					testAccCheckResourceGroupTypeAttribute(groupRsName2, "test 2 group"),
					resource.TestCheckResourceAttr(groupRsName2, "description", "test 2 group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName2, map[string]string{
						"role":    "PROJECT_ADMIN",
						"group":   "test2",
						"project": "test",
					}),
				),
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
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
		},
	})
}

// Test creating a project and adding two different namespace roles
func TestAccParalusResourceProject_Add2GroupsDifferentNamespaceRoles(t *testing.T) {

	groupRsName1 := "paralus_group.test1"
	groupRsName2 := "paralus_group.test2"
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
				resource "paralus_group" "test1" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
				}

				resource "paralus_group" "test2" {
					provider = paralus.valid_resource
					name = "test2"
					description = "test 2 group"
				}

				resource "paralus_project" "add_to_group" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
					project_roles {
						role = "NAMESPACE_READ_ONLY"
						namespace = "platform"
						group = paralus_group.test1.name
					}
					project_roles {
						role = "NAMESPACE_ADMIN"
						namespace = "platform"
						group = paralus_group.test2.name
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "test project"),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"role": "NAMESPACE_READ_ONLY"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"namespace": "platform"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test1"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"role": "NAMESPACE_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"namespace": "platform"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test2"}),

					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupTypeAttribute(groupRsName1, "test 1 group"),
					resource.TestCheckResourceAttr(groupRsName1, "description", "test 1 group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName1, map[string]string{
						"role":      "NAMESPACE_READ_ONLY",
						"group":     "test1",
						"project":   "test",
						"namespace": "platform",
					}),

					testAccCheckResourceGroupExists(groupRsName2),
					testAccCheckResourceGroupTypeAttribute(groupRsName2, "test 2 group"),
					resource.TestCheckResourceAttr(groupRsName2, "description", "test 2 group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName2, map[string]string{
						"role":      "NAMESPACE_ADMIN",
						"group":     "test2",
						"project":   "test",
						"namespace": "platform",
					}),
				),
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
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
		},
	})
}

// Test creating a project and adding two different project roles
func TestAccParalusResourceProject_Add2GroupsDifferentProjectRoles(t *testing.T) {

	groupRsName1 := "paralus_group.test1"
	groupRsName2 := "paralus_group.test2"
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
				resource "paralus_group" "test1" {
					provider = paralus.valid_resource
					name = "test1"
					description = "test 1 group"
				}

				resource "paralus_group" "test2" {
					provider = paralus.valid_resource
					name = "test2"
					description = "test 2 group"
				}

				resource "paralus_project" "add_to_group" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
					project_roles {
						role = "PROJECT_READ_ONLY"
						group = paralus_group.test1.name
					}
					project_roles {
						role = "PROJECT_ADMIN"
						group = paralus_group.test2.name
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "test project"),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"role": "PROJECT_READ_ONLY"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test1"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"project": "test"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"role": "PROJECT_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "project_roles.*", map[string]string{"group": "test2"}),

					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupTypeAttribute(groupRsName1, "test 1 group"),
					resource.TestCheckResourceAttr(groupRsName1, "description", "test 1 group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName1, map[string]string{
						"role":    "PROJECT_READ_ONLY",
						"group":   "test1",
						"project": "test",
					}),

					testAccCheckResourceGroupExists(groupRsName2),
					testAccCheckResourceGroupTypeAttribute(groupRsName2, "test 2 group"),
					resource.TestCheckResourceAttr(groupRsName2, "description", "test 2 group"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName2, map[string]string{
						"role":    "PROJECT_ADMIN",
						"group":   "test2",
						"project": "test",
					}),
				),
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
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
		},
	})
}

// Test creating a project and adding groups with same namespace roles
func TestAccParalusResourceProject_Add2GroupsSameNamespaceRoles(t *testing.T) {

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
						namespace = "platform"
						role = "NAMESPACE_READ_ONLY"
						group = paralus_group.test1.name
					}
					project_roles {
						namespace = "platform"
						role = "NAMESPACE_READ_ONLY"
						group = paralus_group.test2.name
					}
				}`),
				ExpectError: regexp.MustCompile(".*roles must be distinct between project_roles blocks.*"),
			},
		},
	})
}

// Test creating project and adding two groups with same project role
func TestAccParalusResourceProject_Add2GroupsSameProjectRoles(t *testing.T) {

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
						role = "PROJECT_READ_ONLY"
						group = paralus_group.test1.name
					}
					project_roles {
						role = "PROJECT_READ_ONLY"
						group = paralus_group.test2.name
					}
				}`),
				ExpectError: regexp.MustCompile(".*roles must be distinct between project_roles blocks.*"),
			},
		},
	})
}

// Test creating a project and adding two different namespace roles
func TestAccParalusResourceProject_Add2UserRoles(t *testing.T) {

	projectRsName := "paralus_project.add_to_user"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				// we will have a non-empty plan because the access removal will affect the user roles as well
				ExpectNonEmptyPlan: true,
				Config: testAccProviderValidResource(`
				resource "paralus_project" "add_to_user" {
					provider = paralus.valid_resource
					name = "test"
					description = "test project"
					user_roles {
						role = "NAMESPACE_READ_ONLY"
						user = "acctest-user@example.com"
						namespace = "platform"
					}
					user_roles {
						role = "PROJECT_READ_ONLY"
						user = "acctest2-user@example.com"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "test project"),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "user_roles.*", map[string]string{"user": "acctest-user@example.com"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "user_roles.*", map[string]string{"role": "NAMESPACE_READ_ONLY"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "user_roles.*", map[string]string{"namespace": "platform"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "user_roles.*", map[string]string{"user": "acctest2-user@example.com"}),
					resource.TestCheckTypeSetElemNestedAttrs(projectRsName, "user_roles.*", map[string]string{"role": "PROJECT_READ_ONLY"}),
					testAccCheckResourceProjectUserRoleMap(projectRsName, map[string]string{
						"role":      "NAMESPACE_READ_ONLY",
						"user":      "acctest-user@example.com",
						"namespace": "platform",
					}),
					testAccCheckResourceProjectUserRoleMap(projectRsName, map[string]string{
						"role": "PROJECT_READ_ONLY",
						"user": "acctest-user@example.com",
					}),
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

// Verifies project role list has the expected value
func testAccCheckResourceProjectProjectRoleMap(resourceName string, projectRoles map[string]string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		projectStr := rs.Primary.Attributes["name"]

		projectStruct, err := utils.GetProjectByName(projectStr)

		if err != nil {
			return err
		}

		return utils.ValidateProjectNamespaceRolesSet(projectStruct.Spec.ProjectNamespaceRoles, projectRoles)
	}
}

// Verifies user role list has the expected value
func testAccCheckResourceProjectUserRoleMap(resourceName string, userRoles map[string]string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		projectStr := rs.Primary.Attributes["name"]

		projectStruct, err := utils.GetProjectByName(projectStr)

		if err != nil {
			return err
		}

		return utils.ValidateUserRolesSet(projectStruct.Spec.UserRoles, userRoles)
	}
}
