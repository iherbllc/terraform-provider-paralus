package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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

	return clusterStruct
}

// Build the schema resource from Cluster Struct
func BuildResourceFromClusterStruct(cluster *infrav3.Cluster, d *schema.ResourceData) {
	d.Set("name", cluster.Metadata.Name)
	d.Set("description", cluster.Metadata.Description)
	d.Set("project", cluster.Metadata.Project)
	d.Set("cluster_type", cluster.Spec.ClusterType)
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
