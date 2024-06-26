// Group Resource acceptance test
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

// Test missing group name
func TestAccParalusResourceMissingGroup_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

// // Test fail create group if organization name not same as UI configuration
// func TestAccParalusResourceGroupBadOrg_basic(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		// CheckDestroy: testAccCheckGroupResourceDestroy(t),
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccGroupResourceConfigBadOrg(),
// 				ExpectError: regexp.MustCompile(".*not authorized to perform action.*"),
// 			},
// 		},
// 	})
// }

// func testAccGroupResourceConfigBadOrg() string {

// 	conf = paralusProviderConfig()
// 	conf.Organization = "blah"

// 	providerConfig := providerString(conf, "group_badorg_test")
// 	return fmt.Sprintf(`
// 		%s

// 		resource "paralus_group" "badorg_test" {
// 			provider = paralus.group_badorg_test
// 			name = "badorg_group"
// 		}
// 	`, providerConfig)
// }

// General Paralus group resource creation
func TestAccParalusResourceGroup_basic(t *testing.T) {

	groupRsName := "paralus_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "gb-test"
					description = "test group"
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName),
					testAccCheckResourceGroupDescriptionAttribute(groupRsName, "test group"),
					testAccCheckResourceGroupTypeAttribute(groupRsName, "SYSTEM"),
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

			_, err := utils.GetGroupByName(context.Background(), groupStr, nil)

			if err == nil || err != utils.ErrResourceNotExists {
				return fmt.Errorf("group %s still exists", groupStr)
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

		_, err := utils.GetGroupByName(context.Background(), groupStr, nil)

		if err != nil {
			return err
		}
		// log.Printf("group info %s", group)
		return nil
	}
}

// Verifies group attribute is set correctly by Terraform
func testAccCheckResourceGroupDescriptionAttribute(resourceName string, description string) func(s *terraform.State) error {

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

// Verifies group type attribute is set correctly by Terraform
func testAccCheckResourceGroupTypeAttribute(resourceName string, typeStr string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.Attributes["type"] != typeStr {
			return fmt.Errorf("Invalid type")
		}

		return nil
	}
}

// Paralus group creation for one project
func TestAccParalusResourceGroup_Project(t *testing.T) {

	groupRsName := "paralus_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "gp-test"
					description = "test group"
					project_roles {
						role = "ADMIN"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName),
					testAccCheckResourceGroupDescriptionAttribute(groupRsName, "test group"),
					testAccCheckResourceGroupTypeAttribute(groupRsName, "SYSTEM"),
					testAccCheckResourceGroupProjectRoleMap(groupRsName, map[string]string{
						"role":  "ADMIN",
						"group": "gp-test",
					}),
					resource.TestCheckResourceAttr(groupRsName, "description", "test group"),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName, "project_roles.*", map[string]string{"role": "ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName, "project_roles.*", map[string]string{"group": "gp-test"}),
				),
			},
			{
				ResourceName:            groupRsName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_roles"},
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

		group, err := utils.GetGroupByName(context.Background(), groupStr, nil)

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
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
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
					name = "ga2p-test1"
					description = "test 1 group"
					project_roles {
						project = paralus_project.temp.name
						role = "PROJECT_ADMIN"
					}
				}
				resource "paralus_group" "test2" {
					provider = paralus.valid_resource
					name = "ga2p-test2"
					description = "test 2 group"
					project_roles {
						project = paralus_project.temp.name
						role = "PROJECT_ADMIN"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupDescriptionAttribute(groupRsName1, "test 1 group"),
					testAccCheckResourceGroupTypeAttribute(groupRsName1, "SYSTEM"),
					resource.TestCheckResourceAttr(groupRsName1, "description", "test 1 group"),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName1, "project_roles.*", map[string]string{"project": "tempproject"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName1, "project_roles.*", map[string]string{"role": "PROJECT_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName1, "project_roles.*", map[string]string{"group": "ga2p-test1"}),

					testAccCheckResourceGroupExists(groupRsName2),
					testAccCheckResourceGroupDescriptionAttribute(groupRsName2, "test 2 group"),
					testAccCheckResourceGroupTypeAttribute(groupRsName2, "SYSTEM"),
					resource.TestCheckResourceAttr(groupRsName2, "description", "test 2 group"),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName2, "project_roles.*", map[string]string{"project": "tempproject"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName2, "project_roles.*", map[string]string{"role": "PROJECT_ADMIN"}),
					resource.TestCheckTypeSetElemNestedAttrs(groupRsName2, "project_roles.*", map[string]string{"group": "ga2p-test2"}),

					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "A temporary project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "A temporary project"),
					testAccCheckResourceProjectProjectRoleMap(projectRsName, map[string]string{
						"role":    "PROJECT_ADMIN",
						"group":   "ga2p-test1",
						"project": "tempproject",
					}),
					testAccCheckResourceProjectProjectRoleMap(projectRsName, map[string]string{
						"role":    "PROJECT_ADMIN",
						"group":   "ga2p-test2",
						"project": "tempproject",
					}),
				),
			},
			{
				ResourceName:            groupRsName1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_roles"},
			},
			{
				ResourceName:            groupRsName2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_roles"},
			},
			{
				ResourceName:            projectRsName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_roles"},
			},
		},
	})
}

// Create a group with a user
func TestAccParalusResourceGroup_AddUser(t *testing.T) {
	groupRsName1 := "paralus_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "gau-test1"
					description = "test 1 group"
					users = ["acctest-user@example.com"]
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName1),
					testAccCheckResourceGroupDescriptionAttribute(groupRsName1, "test 1 group"),
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
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "ganeu-test1"
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
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "ganep-test1"
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
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "test" {
					provider = paralus.valid_resource
					name = "pnps-test1"
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

		group, err := utils.GetGroupByName(context.Background(), groupStr, nil)

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

// Test multinamespace groups with same namespace
func TestAccParalusNamespaceGroups_Multi(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_group" "namespace_read" {
					provider = paralus.valid_resource
					description = "catalog namespace read group"
					name        = "Catalog Namespace Read"
					type        = "SYSTEM"
					users       = [
						"acctest2-user@example.com",
						"acctest-user@example.com",
					  ]
			  
					project_roles {
						namespace = "catalog"
						project   = "acctest-donotdelete"
						role      = "NAMESPACE_READ_ONLY"
					  }
					project_roles {
						namespace = "garden"
						project   = "acctest-donotdelete"
						role      = "NAMESPACE_READ_ONLY"
					  }
					project_roles {
						namespace = "telemetry"
						project   = "acctest-donotdelete"
						role      = "NAMESPACE_READ_ONLY"
					  }
					project_roles {
						namespace = "garden"
						project   = "acctest-donotdelete"
						role      = "NAMESPACE_READ_ONLY"
					  }
					project_roles {
						namespace = "web"
						project   = "acctest-donotdelete"
						role      = "NAMESPACE_READ_ONLY"
					  }
					}`),
				ExpectError: regexp.MustCompile(".*must have a unique combination.*"),
			},
		},
	})
}

// Test multinamespace groups with same namespace
func TestAccParalusNamespaceGroups_MultiDynamic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				locals {
					groups = [
						{
							description =  "catalog namespace read group"
							name =  "Catalog Namespace Read"
							type =  "SYSTEM"
							users =  [ "acctest2-user@example.com", "acctest-user@example.com"]
							project_roles = [
								{
									namespace =  "catalog"
									project =  "acctest-donotdelete"
									role =  "NAMESPACE_READ_ONLY"
								},
								{
									namespace =  "garden"
									project =  "acctest-donotdelete"
									role =  "NAMESPACE_READ_ONLY"
								},
								{
									namespace = "telemetry"
									project =  "acctest-donotdelete"
									role =  "NAMESPACE_READ_ONLY"
								},
								{
									namespace =  "garden"
									project =  "acctest-donotdelete"
									role =  "NAMESPACE_READ_ONLY"
								},
								{
									namespace =  "web"
									project =  "acctest-donotdelete"
									role =  "NAMESPACE_READ_ONLY"
								}
							]
						}
					]
				}
				resource "paralus_group" "namespace_read" {
					for_each = { for group in local.groups : group.name => group }

					name = each.value.name
					description = each.value.description
					users = can(each.value.users) ? each.value.users : []
				  
					dynamic "project_roles" {
					  for_each = each.value.project_roles
					  content {
						role  = project_roles.value.role
						project = project_roles.value.project
						namespace = project_roles.value.namespace
					  }
					}
				}`),
				ExpectError: regexp.MustCompile(".*must have a unique combination.*"),
			},
		},
	})
}
