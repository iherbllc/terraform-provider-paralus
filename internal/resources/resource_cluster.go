package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/constants"
	"github.com/paralus/cli/pkg/rerror"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// / Paralus Resource Cluster
func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"clusterType": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"provisionType": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provisionEnvironment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provisionPackageType": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"environmentProvider": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kubernetesProvider": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"k8s_yamls": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"partner": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

// Import an existing K8S cluster into a designated project
func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	d.SetId(d.Get("name").(string) + d.Get("project").(string))

	auth := m.(*authprofile.Profile)

	return append(createOrUpdateCluster(ctx, d, auth, "POST"), getClusterYAMLs(ctx, d, auth)...)
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateCluster(ctx, d, m.(*authprofile.Profile), "PUT")
}

// Creates a new cluster or updates an existing one
func createOrUpdateCluster(ctx context.Context, d *schema.ResourceData, auth *authprofile.Profile, requestType string) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterStruct := buildClusterStructFromResource(d)

	// first check to make sure the project exists
	uri := fmt.Sprintf("/infra/v3/project/%s", d.Get("project"))
	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Unknown project %s", d.Get("project"))))
	}

	// make a post call to create the cluster entry provided it exists
	resp, err = auth.AuthAndRequest(uri+"/cluster", requestType, clusterStruct)

	if err != nil {
		howFail := "create"
		if requestType == "PUT" {
			howFail = "update"
		}
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
				d.Get("name"), d.Get("project"))))
	}

	// Update resource information from updated cluster
	if err := buildResourceFromClusterString(resp, d); err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to convert cluster string %s to resource", resp)))
	}

	return diags
}

// Retreive cluster JSON info
func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	auth := m.(*authprofile.Profile)

	// first try using the name filter
	cluster, err := getClusterFast(auth, d)
	if err == nil {
		if err := buildResourceFromClusterString(cluster, d); err == nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to build resource from get response: %s", cluster)))
		}
		return diags
	}

	// get list of clusters
	c, err := listAllClusters(auth, d.Get("project").(string))
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to retrieve all clusters")))
	}

	for _, a := range c {
		if a.Metadata.Name == d.Get("name") {
			// Update resource information from updated cluster
			buildResourceFromClusterStruct(a, d)
			return diags
		}
	}

	d.SetId("")
	return diag.FromErr(errors.Wrap(err,
		fmt.Sprintf("Failed to locate cluster %s in project %s", d.Get("name"), d.Get("project"))))

}

// Looks directly for a cluster based on info provided
func getClusterFast(auth *authprofile.Profile, d *schema.ResourceData) (string, error) {

	if auth == nil {
		auth = config.GetConfig().GetAppAuthProfile()
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", d.Get("project"), d.Get("name"))
	return auth.AuthAndRequest(uri, "GET", nil)

}

// retrieve all clusters from paralus
func listAllClusters(auth *authprofile.Profile, projectId string) ([]*infrav3.Cluster, error) {
	var clusters []*infrav3.Cluster
	limit := 10000
	c, count, err := listClusters(auth, projectId, limit, 0)
	if err != nil {
		return nil, err
	}
	clusters = c
	for count > limit {
		offset := limit
		limit = count
		c, _, err = listClusters(auth, projectId, limit, offset)
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

// Delete an existing cluster
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	resourceClusterRead(ctx, d, m)

	auth := m.(*authprofile.Profile)
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", d.Get("project"), d.Get("name"))
	resp, err := auth.AuthAndRequest(uri, "DELETE", nil)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to delete cluster %s in project %s", d.Get("name"), d.Get("project"))))
	}
	buildClusterStructFromClusterString(resp)

	d.SetId("")
	return diags
}

// Retrieve the YAML files that will be used to setup paralus agents in cluster
func getClusterYAMLs(ctx context.Context, d *schema.ResourceData, auth *authprofile.Profile) diag.Diagnostics {

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s/download", d.Get("project"), d.Get("name"))
	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to retrieve K8s YAMLs for cluster %s in project %s", d.Get("name"), d.Get("project"))))
	}

	d.Set("k8s_yamls", resp)
	return nil
}

// Build a cluster struct from a string
func buildClusterStructFromClusterString(cluster string) (*infrav3.Cluster, error) {
	// Need to take json cluster and convert to the new version
	clusterBytes := []byte(cluster)
	clusterStruct := infrav3.Cluster{}
	if err := json.Unmarshal(clusterBytes, &clusterStruct); err != nil {
		return nil, err
	}
	return &clusterStruct, nil
}

// Build the cluster struct from a schema resource
func buildClusterStructFromResource(d *schema.ResourceData) *infrav3.Cluster {
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
func buildResourceFromClusterString(cluster string, d *schema.ResourceData) error {
	// Need to take json cluster and convert to the new version
	clusterBytes := []byte(cluster)
	clusterStruct := infrav3.Cluster{}
	if err := json.Unmarshal(clusterBytes, &clusterStruct); err != nil {
		return err
	}

	buildResourceFromClusterStruct(&clusterStruct, d)

	return nil
}

// Build the schema resource from Cluster Struct
func buildResourceFromClusterStruct(cluster *infrav3.Cluster, d *schema.ResourceData) {
	d.Set("name", cluster.Metadata.Name)
	d.Set("description", cluster.Metadata.Description)
	d.Set("project", cluster.Metadata.Project)
	d.Set("environmentProvider", cluster.Spec.Params.EnvironmentProvider)
	d.Set("kubernetesProvider", cluster.Spec.Params.KubernetesProvider)
	d.Set("provisionEnvironment", cluster.Spec.Params.ProvisionEnvironment)
	d.Set("provisionPackageType", cluster.Spec.Params.ProvisionPackageType)
	d.Set("provisionType", cluster.Spec.Params.ProvisionType)
	d.Set("state", cluster.Spec.Params.State)
	d.Set("labels", cluster.Metadata.Labels)
	d.Set("annotations", cluster.Metadata.Annotations)
}
