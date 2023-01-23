package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/rerror"
)

// Looks directly for a cluster based on info provided
func GetClusterFast(ctx context.Context, auth *authprofile.Profile,
	project string, cluster string) (string, error) {

	if auth == nil {
		auth = config.GetConfig().GetAppAuthProfile()
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", project, cluster)

	tflog.Trace(ctx, "Cluster GET API Request", map[string]interface{}{
		"uri":    uri,
		"method": "GET",
	})

	return auth.AuthAndRequest(uri, "GET", nil)

}

// retrieve all clusters from paralus
func ListAllClusters(ctx context.Context, auth *authprofile.Profile, project string) ([]*infrav3.Cluster, error) {
	var clusters []*infrav3.Cluster
	limit := 10000
	c, count, err := listClusters(ctx, auth, project, limit, 0)
	if err != nil {
		return nil, err
	}
	clusters = c
	for count > limit {
		offset := limit
		limit = count
		c, _, err = listClusters(ctx, auth, project, limit, offset)
		if err != nil {
			return clusters, err
		}
		clusters = append(clusters, c...)
	}
	return clusters, nil
}

// build a list of all clusters
func listClusters(ctx context.Context, auth *authprofile.Profile,
	project string, limit, offset int) ([]*infrav3.Cluster, int, error) {
	// check to make sure the limit or offset is not negative
	if limit < 0 || offset < 0 {
		return nil, 0, fmt.Errorf("provided limit (%d) or offset (%d) cannot be negative", limit, offset)
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster?limit=%d&offset=%d", project, limit, offset)

	tflog.Trace(ctx, "All Clusters GET API Request", map[string]interface{}{
		"uri":    uri,
		"method": "GET",
	})

	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return nil, 0, rerror.CrudErr{
			Type: "cluster",
			Name: "",
			Op:   "list",
		}
	}

	resp_interf, err := JsonToMap(resp)

	if err != nil {
		return nil, 0, err
	}

	tflog.Trace(ctx, "All Clusters GET API Request", resp_interf)

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

// Build a cluster struct from a resource
func BuildStringFromClusterStruct(cluster *infrav3.Cluster) (string, error) {
	clusterBytes, err := json.Marshal(&cluster)
	if err != nil {
		return "", err
	}

	return string(clusterBytes), nil
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
			ClusterType: d.Get("cluster_type").(string),
		},
	}

	// If we have params, let's add them into the struct
	if params, ok := d.GetOk("params"); ok {
		clusterSet := params.(*schema.Set).List()
		for _, cluster := range clusterSet {
			cluster_params, ok := cluster.(map[string]interface{})
			if ok {
				provisionParams := &infrav3.ProvisionParams{
					EnvironmentProvider:  cluster_params["environment_provider"].(string),
					KubernetesProvider:   cluster_params["kubernetes_provider"].(string),
					ProvisionEnvironment: cluster_params["provision_environment"].(string),
					ProvisionPackageType: cluster_params["provision_package_type"].(string),
					ProvisionType:        cluster_params["provision_type"].(string),
					State:                cluster_params["state"].(string),
				}

				clusterStruct.Spec.Params = provisionParams
			}
		}
	}

	if labels, ok := d.GetOk("labels"); ok {
		clusterStruct.Metadata.Labels = labels.(map[string]string)
	}

	if annotations, ok := d.GetOk("annotations"); ok {
		clusterStruct.Metadata.Annotations = annotations.(map[string]string)
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
	params := map[string]interface{}{
		"environment_provider":   cluster.Spec.Params.EnvironmentProvider,
		"kubernetes_provider":    cluster.Spec.Params.KubernetesProvider,
		"provision_environment":  cluster.Spec.Params.ProvisionEnvironment,
		"provision_package_type": cluster.Spec.Params.ProvisionPackageType,
		"provision_type":         cluster.Spec.Params.ProvisionType,
		"state":                  cluster.Spec.Params.State,
	}
	d.Set("params", params)
	d.Set("labels", cluster.Metadata.Labels)
	d.Set("annotations", cluster.Metadata.Annotations)
}
