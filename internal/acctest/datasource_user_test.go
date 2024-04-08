// Package DataSource project acceptance test
package acctest

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Test empty project name
func TestAccParalusDataSourceNoUsersReturned_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckProjectDataSourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				data "paralus_users" "retrieve_no_users" {
					provider = paralus.valid_resource
					filters {
						email = "blah@blah.com"
					}
				}
			`),
				ExpectError: regexp.MustCompile(".*No users returned based.*"),
			},
		},
	})
}

func TestAccParalusDataSourceMoreThan1QFilterSpecified_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckProjectDataSourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				data "paralus_users" "retrieve_no_users" {
					provider = paralus.valid_resource
					filters {
						email = "blah@blah.com"
						first_name = "blah"
					}
				}
			`),
				ExpectError: regexp.MustCompile(".*specify only one.*"),
			},
		},
	})
}

func TestAccParalusDataSourceBadEmailFilter_basic(t *testing.T) {

	resource_name := "users"
	email := "blah.com"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						email = "%s"
					}
				}`, resource_name, email)),
				ExpectError: regexp.MustCompile(".*must be in format: XXXX@XXX.XXX.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceFilterUsersEmail_basic(t *testing.T) {

	resource_name := "users"
	email := "local-user@example.com"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						email = "%s"
					}
				}`, resource_name, email)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", "", email, false, "", "", ""),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.email", email),
				),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceNoFilteredUserFoundCaseSensitive_basic(t *testing.T) {

	resource_name := "users"
	fname := "local"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						first_name = "%s"
						case_sensitive = true
					}
				}`, resource_name, fname)),
				ExpectError: regexp.MustCompile(".*no user was found using the specified filter.*"),
			},
		},
	})
}

// Standard acceptance test
func TestAccParalusDataSourceFirstName_basic(t *testing.T) {

	resource_name := "users"
	fname := "Local"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						first_name = "%s"
					}
				}`, resource_name, fname)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", fname, "", true, "", "", ""),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.first_name", fname),
				),
			},
		},
	})
}

func TestAccParalusDataSourceFilterUsersLastNameGT1Allowed_basic(t *testing.T) {

	resource_name := "users"
	lname := "User"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						last_name = "%s"
						allow_more_than_one = true
					}
				}`, resource_name, lname)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", "", lname, false, "", "", ""),
					resource.TestCheckTypeSetElemAttr("data.paralus_users."+resource_name, "users_info.*", "2"),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.last_name", lname),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.1.last_name", lname),
				),
			},
		},
	})
}

func TestAccParalusDataSourceFilterUsersFirstName_basic(t *testing.T) {

	resource_name := "users"
	fname := "Local"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						first_name = "%s"
					}
				}`, resource_name, fname)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", fname, "", false, "", "", ""),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.first_name", fname),
				),
			},
		},
	})
}

func TestAccParalusDataSourceFilterByProject_basic(t *testing.T) {

	resource_name := "users"
	project := "default"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						project = "%s"
					}
				}`, resource_name, project)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", "", "", false, project, "", ""),
					resource.TestCheckTypeSetElemAttr("data.paralus_users."+resource_name, "users_info.*", "3"),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.1.project_roles.0.project", project),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.1.project_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccParalusDataSourceFilterByGroup_basic(t *testing.T) {

	resource_name := "users"
	group := "acctest-group"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						group = "%s"
					}
				}`, resource_name, group)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", "", "", false, "", "", group),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.groups.1", group),
					resource.TestCheckTypeSetElemAttr("data.paralus_users."+resource_name, "users_info.*", "1"),
					resource.TestCheckTypeSetElemAttr("data.paralus_users."+resource_name, "users_info.0.groups.*", "2"),
				),
			},
		},
	})
}

func TestAccParalusDataSourceFilterByRole_basic(t *testing.T) {

	resource_name := "users"
	role := "NAMESPACE_READ_ONLY"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
					filters {
						role = "%s"
					}
				}`, resource_name, role)),
				Check: resource.ComposeTestCheckFunc(
					testReturnEquals("data.paralus_users."+resource_name, "", "", "", false, "", role, ""),
					resource.TestCheckTypeSetElemAttr("data.paralus_users."+resource_name, "users_info.*", "1"),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.project_roles.0.role", role),
					resource.TestCheckResourceAttr("data.paralus_users."+resource_name, "users_info.0.project_roles.*", "1"),
				),
			},
		},
	})
}

func TestAccParalusDataSourceNoFilter_basic(t *testing.T) {

	resource_name := "users"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(fmt.Sprintf(`
				data "paralus_users" "%s" {
					provider = paralus.valid_resource
				}`, resource_name)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("data.paralus_users."+resource_name, "users_info.*", "5"),
				),
			},
		},
	})
}

// Verifies project attribute is set correctly by Terraform
func testReturnEquals(resourceName string, email string, first_name string, last_name string,
	case_sensitive bool, project string, role string, group string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		if !case_sensitive {
			email = strings.ToLower(email)
			first_name = strings.ToLower(first_name)
			last_name = strings.ToLower(last_name)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if email != "" {
			returned_email := rs.Primary.Attributes["users_info.0.email"]
			if !case_sensitive {
				returned_email = strings.ToLower(returned_email)
			}
			if returned_email != email {
				return fmt.Errorf("invalid email checked: %s", email)
			}
		}
		if first_name != "" {
			returned_first_name := rs.Primary.Attributes["users_info.0.first_name"]
			if !case_sensitive {
				returned_first_name = strings.ToLower(returned_first_name)
			}
			if returned_first_name != first_name {
				return fmt.Errorf("invalid first name checked: %s", first_name)
			}
		}
		if last_name != "" {
			returned_last_name := rs.Primary.Attributes["users_info.0.last_name"]
			if !case_sensitive {
				returned_last_name = strings.ToLower(returned_last_name)
			}
			if returned_last_name != last_name {
				return fmt.Errorf("invalid last name checked: %s", last_name)
			}
		}
		if project != "" {
			returned_project := rs.Primary.Attributes["users_info.1.project_roles.0.project"]
			if returned_project != project {
				return fmt.Errorf("invalid project checked: %s", project)
			}
		}
		if role != "" {
			returned_role := rs.Primary.Attributes["users_info.0.project_roles.0.role"]
			if returned_role != role {
				return fmt.Errorf("invalid role checked: %s", role)
			}
		}
		if group != "" {
			returned_group := rs.Primary.Attributes["users_info.0.groups.1"]
			if returned_group != group {
				return fmt.Errorf("invalid group checked: %s", group)
			}
		}
		return nil
	}
}
