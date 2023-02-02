// Group Resource acceptance test
package acctest

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/group"
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
				ExpectError: regexp.MustCompile(".*name cannot be empty.*"),
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
				ExpectError: regexp.MustCompile(".*could not complete operation.*"),
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

// testAccCheckGroupResourceDestroy verifies the cluster has been destroyed
func testAccCheckGroupResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_group" {
				continue
			}

			groupStr := rs.Primary.Attributes["name"]

			_, err := group.GetGroupByName(groupStr)

			if err == nil {
				group.DeleteGroup(groupStr)
			}
		}

		return nil
	}
}

// testAccCheckGroupExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Group instance
func testAccCheckResourceGroupExists(resourceName string) func(s *terraform.State) error {

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

		group, err := group.GetGroupByName(groupStr)

		if err == nil {
			return err
		}
		log.Printf("group info %s", group)
		return nil
	}
}

// testAccCheckGroupTypeAttribute verifies group attribute is set correctly by
// Terraform
func testAccCheckResourceGroupTypeAttribute(resourceName string, description string) func(s *terraform.State) error {

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

// Paralus group creation for all projects admin rights
func TestAccParalusResourceGroup_AlProjects(t *testing.T) {

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
						group = "test"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(groupRsName),
					testAccCheckResourceGroupTypeAttribute(groupRsName, "test group"),
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
