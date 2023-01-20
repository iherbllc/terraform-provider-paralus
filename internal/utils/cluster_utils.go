package utils

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/constants"
	"github.com/paralus/cli/pkg/rerror"
)

// Looks directly for a cluster based on info provided
func GetClusterFast(auth *authprofile.Profile, project string, cluster string) (string, error) {

	if auth == nil {
		auth = config.GetConfig().GetAppAuthProfile()
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", project, cluster)
	return auth.AuthAndRequest(uri, "GET", nil)

}

// retrieve all clusters from paralus
func ListAllClusters(auth *authprofile.Profile, project string) ([]*infrav3.Cluster, error) {
	var clusters []*infrav3.Cluster
	limit := 10000
	c, count, err := listClusters(auth, project, limit, 0)
	if err != nil {
		return nil, err
	}
	clusters = c
	for count > limit {
		offset := limit
		limit = count
		c, _, err = listClusters(auth, project, limit, offset)
		if err != nil {
			return clusters, err
		}
		clusters = append(clusters, c...)
	}
	return clusters, nil
}

// build a list of all clusters
func listClusters(auth *authprofile.Profile, project string, limit, offset int) ([]*infrav3.Cluster, int, error) {
	// check to make sure the limit or offset is not negative
	if limit < 0 || offset < 0 {
		return nil, 0, fmt.Errorf("provided limit (%d) or offset (%d) cannot be negative", limit, offset)
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster?limit=%d&offset=%d", project, limit, offset)
	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return nil, 0, rerror.CrudErr{
			Type: "cluster",
			Name: "",
			Op:   "list",
		}
	}
	a := infrav3.ClusterList{}

	if err := json.Unmarshal([]byte(resp), &a); err != nil {
		return nil, 0, err
	}

	return a.Items, int(a.Metadata.Count), nil
}

// Build a cluster struct from a resource
func BuildClusterStructFromString(clusterStr string, cluster *infrav3.Cluster) error {
	// Need to take json cluster and convert to the new version
	clusterBytes := []byte(clusterStr)
	if err := json.Unmarshal(clusterBytes, &cluster); err != nil {
		return err
	}

	return nil
}

// Build the cluster struct from a schema resource
func BuildClusterStructFromResource(d *schema.ResourceData) *infrav3.Cluster {
	clusterStruct := infrav3.Cluster{
		Kind: "Cluster",
		Metadata: &commonv3.Metadata{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Project:     d.Get("project").(string),
		},
		Spec: &infrav3.ClusterSpec{
			Metro:       &infrav3.Metro{},
			ClusterType: constants.CLUSTER_TYPE_IMPORT,
			Params: &infrav3.ProvisionParams{
				EnvironmentProvider:  d.Get("environmentProvider").(string),
				KubernetesProvider:   d.Get("kubernetesProvider").(string),
				ProvisionEnvironment: d.Get("provisionEnvironment").(string),
				ProvisionPackageType: d.Get("provisionPackageType").(string),
				ProvisionType:        d.Get("provisionType").(string),
				State:                d.Get("state").(string),
			},
		},
	}

	if d.Get("labels") != nil {
		clusterStruct.Metadata.Labels = d.Get("labels").(map[string]string)
	}

	if d.Get("annotations") != nil {
		clusterStruct.Metadata.Annotations = d.Get("annotations").(map[string]string)
	}

	return &clusterStruct
}

// Build a resource from a cluster struct
func BuildResourceFromClusterString(cluster string, d *schema.ResourceData) error {
	// Need to take json cluster and convert to the new version
	clusterBytes := []byte(cluster)
	clusterStruct := infrav3.Cluster{}
	if err := json.Unmarshal(clusterBytes, &clusterStruct); err != nil {
		return err
	}

	BuildResourceFromClusterStruct(&clusterStruct, d)

	return nil
}

// Build the schema resource from Cluster Struct
func BuildResourceFromClusterStruct(cluster *infrav3.Cluster, d *schema.ResourceData) {
	d.Set("name", cluster.Metadata.Name)
	d.Set("description", cluster.Metadata.Description)
	d.Set("project", cluster.Metadata.Project)
	d.Set("environment_provider", cluster.Spec.Params.EnvironmentProvider)
	d.Set("kubernetes_provider", cluster.Spec.Params.KubernetesProvider)
	d.Set("provision_environment", cluster.Spec.Params.ProvisionEnvironment)
	d.Set("provision_package_type", cluster.Spec.Params.ProvisionPackageType)
	d.Set("provision_type", cluster.Spec.Params.ProvisionType)
	d.Set("state", cluster.Spec.Params.State)
	d.Set("labels", cluster.Metadata.Labels)
	d.Set("annotations", cluster.Metadata.Annotations)
}
