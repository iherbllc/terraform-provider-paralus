// Cluster Resource acceptance test
package acctest

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
)

// Test missing cluster name
func TestAccParalusResourceMissingCluster_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterResourceConfigMissingCluster(),
				ExpectError: regexp.MustCompile(".*argument \"name\" is required.*"),
			},
		},
	})
}

func testAccClusterResourceConfigMissingCluster() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "cluster_missing_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_cluster" "missingname_test" {
			provider = paralus.cluster_missing_name
		}
	`, providerConfig)
}

// Test missing project name
func TestAccParalusResourceClusterMissingProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterResourceConfigMissingProject(),
				ExpectError: regexp.MustCompile(".*argument \"project\" is required.*"),
			},
		},
	})
}

func testAccClusterResourceConfigMissingProject() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "project_missing_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_cluster" "missingname_test" {
			provider = paralus.project_missing_name
			name = "crcmp-test"
		}
	`, providerConfig)
}

// Test empty project name
func TestAccParalusResourceClusterEmptyProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterResourceConfiEmptyProject(),
				ExpectError: regexp.MustCompile(".*expected not empty string.*"),
			},
		},
	})
}

func testAccClusterResourceConfiEmptyProject() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "project_empty_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_cluster" "missingname_test" {
			provider = paralus.project_empty_name
			name = "crcep-test"
			project = ""
			cluster_type = "imported"
		}
	`, providerConfig)
}

// Test empty cluster name
func TestAccParalusResourceEmptyCluster_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterResourceConfigEmptyCluster(),
				ExpectError: regexp.MustCompile(".*expected not empty string.*"),
			},
		},
	})
}

func testAccClusterResourceConfigEmptyCluster() string {

	conf = paralusProviderConfig()
	providerConfig := providerString(conf, "cluster_empty_name")
	return fmt.Sprintf(`
		%s

		resource "paralus_cluster" "emptyname_test" {
			provider = paralus.cluster_empty_name
			name = ""
			project = "ccec-test"
			cluster_type = "imported"
		}
	`, providerConfig)
}

// Test project and cluster creation
func TestAccParalusResourceProjectCluster_full(t *testing.T) {

	projectRsName := "paralus_project.testproject"
	clusterRsName := "paralus_cluster.testcluster"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_project" "testproject" {
					provider = paralus.valid_resource
					name = "projectresource4"
					description = "from acct test"
				}
		
				resource "paralus_cluster" "testcluster" {
					provider = paralus.valid_resource
					name = "clusterresource4"
					project = paralus_project.testproject.name
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}
				`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceClusterExists(clusterRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "from acct test"),
					testAccCheckResourceClusterTypeAttribute(clusterRsName, "imported"),
					testAccCheckResourceAttributeSet(clusterRsName, "relays"),
					resource.TestCheckResourceAttr(projectRsName, "description", "from acct test"),
					resource.TestCheckResourceAttr(clusterRsName, "project", "projectresource4"),
					resource.TestCheckTypeSetElemAttr(clusterRsName, "bootstrap_files.*", "12"),
				),
			},
			{
				ResourceName:            clusterRsName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"annotations", "labels"},
			},
			{
				ResourceName:      projectRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Test unknown project
func TestAccParalusResourceCluster_MissingClusterType(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "cmct-test"
					project = "acctest-donotdelete"
				}`),
				ExpectError: regexp.MustCompile(".*argument \"cluster_type\" is required.*"),
			},
		},
	})
}

// Test unknown project
func TestAccParalusResourceClusterUnknownProject_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "cup-test"
					project = "blah"
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}`),
				ExpectError: regexp.MustCompile(".*project .* does not exist.*"),
			},
		},
	})
}

// Test cluster creation into existing project, looking up project using datasource
func TestAccParalusResourceCluster_WithProjectDatasource(t *testing.T) {
	clusterRsName := "paralus_cluster.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				data "paralus_project" "test" {
					name = "acctest-donotdelete"
				}
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "pds-cluster-test1"
					project = data.paralus_project.test.name
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceClusterExists(clusterRsName),
					testAccCheckResourceClusterTypeAttribute(clusterRsName, "imported"),
					testAccCheckResourceAttributeSet(clusterRsName, "relays"),
					resource.TestCheckResourceAttr(clusterRsName, "project", "acctest-donotdelete"),
				),
			},
			{
				ResourceName:      clusterRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Test cluster creation into new project
func TestAccParalusResourceCluster_Full(t *testing.T) {
	clusterRsName := "paralus_cluster.test"
	projectRsName := "paralus_project.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_project" "test" {
					name = "dynamic-acctest"
					description = "dynamic test project"
				}
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "dynamic-test1"
					project = paralus_project.test.name
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceClusterExists(clusterRsName),
					testAccCheckResourceClusterTypeAttribute(clusterRsName, "imported"),
					testAccCheckResourceAttributeSet(clusterRsName, "relays"),
					resource.TestCheckResourceAttr(clusterRsName, "project", "dynamic-acctest"),
					testAccCheckResourceProjectExists(projectRsName),
					testAccCheckResourceProjectTypeAttribute(projectRsName, "dynamic test project"),
					resource.TestCheckResourceAttr(projectRsName, "description", "dynamic test project"),
				),
			},
			{
				ResourceName:      clusterRsName,
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

// Test cluster creation into existing project
func TestAccParalusResourceCluster_basic(t *testing.T) {
	clusterRsName := "paralus_cluster.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccConfigPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderValidResource(`
				resource "paralus_cluster" "test" {
					provider = paralus.valid_resource
					name = "basic-test1"
					project = "acctest-donotdelete"
					cluster_type = "imported"
					params {
						provision_type = "IMPORT"
						provision_environment = "CLOUD"
						kubernetes_provider = "EKS"
						state = "PROVISION"
					}
				}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceClusterExists(clusterRsName),
					testAccCheckResourceClusterTypeAttribute(clusterRsName, "imported"),
					testAccCheckResourceAttributeSet(clusterRsName, "relays"),
					resource.TestCheckResourceAttr(clusterRsName, "project", "acctest-donotdelete"),
				),
			},
			{
				ResourceName:      clusterRsName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Verifies the cluster has been destroyed
// and destroys it if it is not
func testAccCheckClusterResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_cluster" {
				continue
			}

			project := rs.Primary.Attributes["project"]
			clusterName := rs.Primary.Attributes["name"]

			_, err := utils.GetCluster(context.Background(), clusterName, project, nil)

			if err == nil || err != utils.ErrResourceNotExists {
				return utils.DeleteCluster(context.Background(), clusterName, project, nil)
			}
		}

		return nil
	}
}

// Uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Cluster instance
func testAccCheckResourceClusterExists(resourceName string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return errors.New("cluster id is not set")
		}

		project := rs.Primary.Attributes["project"]
		clusterName := rs.Primary.Attributes["name"]

		_, err := utils.GetCluster(context.Background(), clusterName, project, nil)

		if err != nil {
			return err
		}
		return nil
	}
}

// Verifies cluster attribute is set correctly by Terraform
func testAccCheckResourceClusterTypeAttribute(resourceName string, cluster_type string) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.Attributes["cluster_type"] != cluster_type {
			return fmt.Errorf("invalid cluster type")
		}

		return nil
	}
}

// Tests that a resource attribute has a value
func testAccCheckResourceAttributeSet(resourceName string, attrName string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.Attributes[attrName] == "" {
			return fmt.Errorf(fmt.Sprintf("attribute %s is empty", attrName))
		}

		return nil
	}
}

func TestResource_UpgradeFromVersion(t *testing.T) {
	/* ... */
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"paralus": {
						VersionConstraint: "0.0.47",
						Source:            "iherbllc/paralus",
					},
				},
				Config: testAccProviderValidResource(`resource "paralus_cluster" "example" {
							provider = paralus.valid_resource
							name = "basic-test1"
							project = "acctest-donotdelete"
							cluster_type = "imported"
							params {
								provision_type = "IMPORT"
								provision_environment = "CLOUD"
								kubernetes_provider = "EKS"
								state = "PROVISION"
							}
                        }`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("paralus_cluster.example", "name", "basic-test1"),
					resource.TestCheckResourceAttr("paralus_cluster.example", "project", "acctest-donotdelete"),
					/* ... */
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Config: testAccProviderValidResource(`resource "paralus_cluster" "example" {
							provider = paralus.valid_resource
							name = "basic-test1"
							project = "acctest-donotdelete"
							cluster_type = "imported"
							params {
								provision_type = "IMPORT"
								provision_environment = "CLOUD"
								kubernetes_provider = "EKS"
								state = "PROVISION"
							}
                        }`),
				// ConfigPlanChecks is a terraform-plugin-testing feature.
				// If acceptance testing is still using terraform-plugin-sdk/v2,
				// use `PlanOnly: true` instead. When migrating to
				// terraform-plugin-testing, switch to `ConfigPlanChecks` or you
				// will likely experience test failures.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
