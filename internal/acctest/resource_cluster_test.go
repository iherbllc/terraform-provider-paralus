package acctest

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/paralus/cli/pkg/authprofile"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"

	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

func TestAccParalusCluster_basic(t *testing.T) {
	var clusterStruct infrav3.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccClusterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("paralus.cluster", &clusterStruct),
					testAccCheckClusterTypeAttribute(&clusterStruct, "imported"),
					resource.TestCheckResourceAttr("paralus.cluster", "project", "test"),
				),
			},
		},
	})
}

func testAccClusterConfig() string {

	config := paralusProviderConfig()
	return fmt.Sprintf(`
		provider "paralus-ctl" {
			version = "1.0"
			alias = "cluster_test"
			profile = "%s"
			rest_endpoint = "%s"
			ops_endpoint = "%s"
			api_key = "%s"
			api_secret = "%s"
		}

		resource "cluster" "test" {
			provider = "paralus-ctl.cluster_test"
			name = "test"
			description = "test cluster"
			project = "default"
			cluster_type = "imported"
			provision_type = "IMPORT"
			provision_environment = "CLOUD"
			kubernetes_provider = "EKS"
			state = "PROVISION"
		}
	`, config.Profile, config.RESTEndpoint, config.OPSEndpoint,
		config.APIKey, config.APISecret)
}

// makes sure necessary PCTL values are set
func testAccClusterPreCheck(t *testing.T) {
	if v := os.Getenv("PCTL_PROFILE"); v == "" {
		t.Fatal("PCTL_PROFILE must be set for acceptance tests")
	}
	if v := os.Getenv("PCTL_REST_ENDPOINT"); v == "" {
		t.Fatal("PCTL_REST_ENDPOINT must be set for acceptance tests")
	}

	if v := os.Getenv("PCTL_OPS_ENDPOINT"); v == "" {
		t.Fatal("PCTL_OPS_ENDPOINT must be set for acceptance tests")
	}

	if v := os.Getenv("PCTL_API_KEY"); v == "" {
		t.Fatal("PCTL_API_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("PCTL_API_SECRET"); v == "" {
		t.Fatal("PCTL_API_SECRET must be set for acceptance tests")
	}

	if v := os.Getenv("PCTL_REST_ENDPOINT"); v == "" {
		t.Fatal("PCTL_REST_ENDPOINT must be set for acceptance tests")
	}
}

// testAccCheckClusterResourceDestroy verifies the cluster has been destroyed
func testAccCheckClusterResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the connection established in Provider configuration
		auth := testAccProvider.Meta().(*authprofile.Profile)

		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "cluster" {
				continue
			}

			project := rs.Primary.Attributes["project"]
			cluster := rs.Primary.ID

			// first try using the name filter
			_, err := paralusUtils.GetClusterFast(auth, project, cluster)

			if err == nil {
				return fmt.Errorf("Cluster (%s) still exists.", rs.Primary.ID)
			}

			// get list of clusters
			clusters, err := paralusUtils.ListAllClusters(auth, project)
			if err != nil {
				return nil
			}

			for _, a := range clusters {
				if a.Metadata.Name == cluster {
					return fmt.Errorf("Cluster (%s) still exists.", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

// testAccCheckClusterExists uses the paralus API through PCTL to retrieve cluster info
// and store it as a PCTL Cluster instance
func testAccCheckClusterExists(resourceName string, cluster *infrav3.Cluster) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Cluster ID is not set")
		}

		// retrieve the connection established in Provider configuration
		auth := testAccProvider.Meta().(*authprofile.Profile)

		project := rs.Primary.Attributes["project"]
		cluster_id := rs.Primary.ID

		// first try using the name filter
		cluster_json, _ := paralusUtils.GetClusterFast(auth, project, cluster_id)

		if cluster_json != "" {
			paralusUtils.BuildClusterStructFromString(cluster_json, cluster)
		}

		// get list of clusters
		clusters, err := paralusUtils.ListAllClusters(auth, project)
		if err != nil {
			return nil
		}

		for _, a := range clusters {
			if a.Metadata.Name == cluster_id {
				return fmt.Errorf("Cluster (%s) still exists.", rs.Primary.ID)
			}
		}

		return nil
	}
}

// testAccCheckClusterTypeAttribute verifies project attribute is set correctly by
// Terraform
func testAccCheckClusterTypeAttribute(cluster *infrav3.Cluster, cluster_type string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		if cluster.Metadata.Project != cluster_type {
			return fmt.Errorf("Cluster Type not set correctly")
		}

		return nil
	}
}
