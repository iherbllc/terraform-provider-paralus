package acctest

import (
	"context"
	"fmt"
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
		PreCheck:     func() { testAccConfigPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterResourceDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("paralusctl.paralus_cluster", &clusterStruct),
					testAccCheckClusterTypeAttribute(&clusterStruct, "imported"),
					resource.TestCheckResourceAttr("paralusctl.paralus_cluster", "project", "default"),
				),
			},
		},
	})
}

func testAccClusterConfig() string {

	providerConfig := providerString(nil, "cluster_test")
	return fmt.Sprintf(`
		%s

		resource "paralus_cluster" "test" {
			provider = "paralusctl.cluster_test"
			name = "test"
			description = "test cluster"
			project = "default"
			cluster_type = "imported"
			params {
				provision_type = "IMPORT"
				provision_environment = "CLOUD"
				kubernetes_provider = "EKS"
				state = "PROVISION"
			}
		}
	`, providerConfig)
}

// testAccCheckClusterResourceDestroy verifies the cluster has been destroyed
func testAccCheckClusterResourceDestroy(t *testing.T) func(s *terraform.State) error {

	return func(s *terraform.State) error {
		// retrieve the connection established in Provider configuration
		auth := testAccProvider.Meta().(*authprofile.Profile)

		// loop through the resources in state, verifying each widget
		// is destroyed
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "paralus_cluster" {
				continue
			}

			project := rs.Primary.Attributes["project"]
			cluster := rs.Primary.Attributes["name"]

			// first try using the name filter
			_, err := paralusUtils.GetClusterFast(context.Background(), auth, project, cluster)

			if err == nil {
				return fmt.Errorf("Cluster (%s) still exists.", rs.Primary.ID)
			}

			// get list of clusters
			clusters, err := paralusUtils.ListAllClusters(context.Background(), auth, project)
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
		cluster_id := rs.Primary.Attributes["name"]

		// first try using the name filter
		cluster_json, _ := paralusUtils.GetClusterFast(context.Background(), auth, project, cluster_id)

		if cluster_json != "" {
			paralusUtils.BuildClusterStructFromString(cluster_json, cluster)
		}

		// get list of clusters
		clusters, err := paralusUtils.ListAllClusters(context.Background(), auth, project)
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
