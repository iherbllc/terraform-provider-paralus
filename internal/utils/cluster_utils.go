// Utility methods for PCTL Cluster manipulation
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8Scheme "k8s.io/client-go/kubernetes/scheme"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

// Build the cluster struct from a schema resource
func BuildClusterStructFromResource(d *schema.ResourceData) *infrav3.Cluster {

	clusterStruct := &infrav3.Cluster{
		Kind: "Cluster",
		Metadata: &commonv3.Metadata{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Project:     d.Get("project").(string),
			Id:          d.Get("uuid").(string),
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
		if clusterStruct.Metadata.Labels == nil {
			clusterStruct.Metadata.Labels = make(map[string]string)
		}
		for k, v := range labels.(map[string]interface{}) {
			clusterStruct.Metadata.Labels[k] = v.(string)
		}
	}

	if annotations, ok := d.GetOk("annotations"); ok {
		if clusterStruct.Metadata.Annotations == nil {
			clusterStruct.Metadata.Annotations = make(map[string]string)
		}
		for k, v := range annotations.(map[string]interface{}) {
			clusterStruct.Metadata.Annotations[k] = v.(string)
		}
	}

	return clusterStruct
}

// Build the schema resource from Cluster Struct
func BuildResourceFromClusterStruct(cluster *infrav3.Cluster, d *schema.ResourceData) {
	d.Set("name", cluster.Metadata.Name)
	d.Set("description", cluster.Metadata.Description)
	d.Set("project", cluster.Metadata.Project)
	d.Set("cluster_type", cluster.Spec.ClusterType)
	d.Set("uuid", cluster.Metadata.Id)
	if cluster.Spec.Params != nil {
		params := d.Get("params").(*schema.Set)
		params.Add(map[string]interface{}{
			"environment_provider":   cluster.Spec.Params.EnvironmentProvider,
			"kubernetes_provider":    cluster.Spec.Params.KubernetesProvider,
			"provision_environment":  cluster.Spec.Params.ProvisionEnvironment,
			"provision_package_type": cluster.Spec.Params.ProvisionPackageType,
			"provision_type":         cluster.Spec.Params.ProvisionType,
			"state":                  cluster.Spec.Params.State,
		})
		d.Set("params", params)
	}
	d.Set("labels", cluster.Metadata.Labels)
	d.Set("annotations", cluster.Metadata.Annotations)
}

// Splits a single YAML file containing multiple YAML entries into a list of string
func splitSingleYAMLIntoList(singleYAML string) []string {
	docs := strings.Split(string(singleYAML), "\n---")

	yamls := []string{}
	// Trim whitespace in both ends of each yaml docs.
	// - Re-add a single newline last
	for _, doc := range docs {
		content := strings.TrimSpace(doc)
		// Ignore empty docs
		if content != "" {
			yamls = append(yamls, content)
		}
	}
	return yamls
}

// Retrieve the relays info from the bootstrap files
// Which are found within the relay-agent-config configmap
func getBootstrapRelays(bootstrapFiles []string) (string, error) {
	// yamlFiles is an []string
	for _, boostrapFile := range bootstrapFiles {

		decode := k8Scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(boostrapFile), nil, nil)

		if err != nil {
			return boostrapFile, err
		}

		// now use switch over the type of the object
		// and match each type-case
		switch o := obj.(type) {
		// case *v1.Pod:
		// 	// o is a pod
		// case *v1beta1.Role:
		// 	// o is the actual role Object with all fields etc
		// case *v1beta1.RoleBinding:
		// case *v1beta1.ClusterRole:
		// case *v1beta1.ClusterRoleBinding:
		// case *v1.ServiceAccount:
		case *v1.ConfigMap:
			targetConfigMap := o.Data
			if relays, ok := targetConfigMap["relays"]; ok {
				return relays, nil
			}
		default:
			//o is unknown for us
		}
	}
	return "", nil
}

// Retrieve the YAML files that will be used to setup paralus agents in cluster and assign it to the schema
// Also retrieve the relays from  the data of the relay-agent configMap YAML
// Due to the parallel nature of testing, it might be that the cluster would be created
// before the relay was effectively populated. So let's do a increased delay check
func SetBootstrapFileAndRelays(ctx context.Context, d *schema.ResourceData) error {

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	b := &backoff.Backoff{
		Jitter: true,
		Max:    5 * time.Minute,
	}
	rand.Seed(42)

	var bootstrapFile string
	var bootstrapFiles []string
	var relay_or_resp string
	var err error

	for {
		// already checked earlier for cluster to exist, so don't have to check again.
		bootstrapFile, err = GetBootstrapFile(clusterId, projectId)

		if err != nil {
			return errors.Wrapf(err, "Error retrieving bootstrap file for cluster %s in project %s",
				clusterId, projectId)
		}

		bootstrapFiles = splitSingleYAMLIntoList(bootstrapFile)
		relay_or_resp, err = getBootstrapRelays(bootstrapFiles)
		if err != nil {
			return errors.Wrapf(err, "Error while decoding YAML object %s", relay_or_resp)
		}

		d := b.Duration()
		if relay_or_resp != "" || d >= b.Max {
			break
		}

		// If the GetBootstrapFile call is too fast, it might lead to the relay info not ending up in the cluster
		// yet, so will need to try again. Using jitter to avoid flooding the API.

		tflog.Info(ctx, fmt.Sprintf("No relay populated yet, retrying in %s", d))
		time.Sleep(d)
	}

	if relay_or_resp == "" {
		return errors.Errorf("Unable to retrieve relay info from created cluster within %s", b.Duration())
	}

	b.Reset()
	d.Set("relays", relay_or_resp)
	d.Set("bootstrap_files", bootstrapFiles)
	d.Set("bootstrap_files_combined", bootstrapFile)

	return nil
}

// Will retrieve the bootstrap file for imported clusters
func GetBootstrapFile(name, project string) (string, error) {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s/download", project, name)
	return makeRestCall(uri, "GET", nil)

}

// Retrieves cluster info
func GetCluster(name, project string) (*infrav3.Cluster, error) {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", project, name)
	resp, err := makeRestCall(uri, "GET", nil)
	if err != nil {
		return nil, err
	}
	var cluster infrav3.Cluster
	if err := json.Unmarshal([]byte(resp), &cluster); err != nil {
		return nil, fmt.Errorf("error unmarshalling cluster details: %s", err)
	}
	return &cluster, nil
}

// Delete the cluster
func DeleteCluster(name, project string) error {
	// get cluster
	_, err := GetCluster(name, project)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", project, name)
	_, err = makeRestCall(uri, "DELETE", nil)
	if err != nil {
		return err
	}

	return nil
}

// Update cluster takes the updated cluster details and sends it to the core
func CreateCluster(cluster *infrav3.Cluster) error {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster", cluster.Metadata.Project)
	_, err := makeRestCall(uri, "POST", cluster)
	if err != nil {
		return err
	}
	return nil
}

// Update cluster takes the updated cluster details and sends it to the core
func UpdateCluster(cluster *infrav3.Cluster) error {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", cluster.Metadata.Project, cluster.Metadata.Name)
	_, err := makeRestCall(uri, "PUT", cluster)
	if err != nil {
		return err
	}
	return nil
}

// ListAllClusters uses the lower level func ListClusters to retrieve a list of all clusters
func ListAllClusters(projectId string) ([]*infrav3.Cluster, error) {
	var clusters []*infrav3.Cluster
	limit := 10000
	c, count, err := listClusters(projectId, limit, 0)
	if err != nil {
		return nil, err
	}
	clusters = c
	for count > limit {
		offset := limit
		limit = count
		c, _, err = listClusters(projectId, limit, offset)
		if err != nil {
			return clusters, err
		}
		clusters = append(clusters, c...)
	}
	return clusters, nil
}

/*
ListClusters paginates through a list of clusters
*/
func listClusters(project string, limit, offset int) ([]*infrav3.Cluster, int, error) {
	// check to make sure the limit or offset is not negative
	if limit < 0 || offset < 0 {
		return nil, 0, fmt.Errorf("provided limit (%d) or offset (%d) cannot be negative", limit, offset)
	}
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster?limit=%d&offset=%d", project, limit, offset)
	resp, err := makeRestCall(uri, "GET", nil)
	if err != nil {
		return nil, 0, err
	}
	a := infrav3.ClusterList{}
	_ = json.Unmarshal([]byte(resp), &a)
	return a.Items, int(a.Metadata.Count), nil
}
