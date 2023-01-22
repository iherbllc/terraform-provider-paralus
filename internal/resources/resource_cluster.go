package resources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/authprofile"

	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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
			"cluster_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"provision_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provision_environment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provision_package_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"environment_provider": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kubernetes_provider": {
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

	clusterStruct := paralusUtils.BuildClusterStructFromResource(d)

	// first check to make sure the project exists
	uri := fmt.Sprintf("/infra/v3/project/%s", d.Get("project"))

	tflog.Trace(ctx, "Project Info API Request", map[string]interface{}{
		"uri":    uri,
		"method": "GET",
	})

	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Unknown project %s", d.Get("project"))))
	}

	resp_interf, err := utils.JsonToMap(resp)

	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed converting project API response to map %s", resp)))
	}

	tflog.Trace(ctx, "Project Info API Response", resp_interf)

	uri = uri + "/cluster"

	clusterStr, _ := paralusUtils.BuildStringFromClusterStruct(clusterStruct)

	tflog.Trace(ctx, "Cluster Info API Request", map[string]interface{}{
		"uri":     uri,
		"method":  requestType,
		"payload": clusterStr,
	})

	// make a post call to create the cluster entry provided it exists
	resp, err = auth.AuthAndRequest(uri, requestType, clusterStruct)

	if err != nil {
		howFail := "create"
		if requestType == "PUT" {
			howFail = "update"
		}
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
				d.Get("name"), d.Get("project"))))
	}

	resp_interf, err = utils.JsonToMap(resp)

	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed converting cluster API response to map %s", resp)))
	}

	tflog.Trace(ctx, "Cluster Info API Response", resp_interf)

	// Update resource information from updated cluster
	if err := paralusUtils.BuildResourceFromClusterString(resp, d); err != nil {
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
	cluster, err := paralusUtils.GetClusterFast(ctx, auth, d.Get("project").(string), d.Get("name").(string))
	if err == nil {
		if err := paralusUtils.BuildResourceFromClusterString(cluster, d); err == nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to build resource from get response: %s", cluster)))
		}
		return diags
	}

	// get list of clusters
	c, err := paralusUtils.ListAllClusters(ctx, auth, d.Get("project").(string))
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "Failed to retrieve all clusters"))
	}

	for _, a := range c {
		if a.Metadata.Name == d.Get("name") {
			// Update resource information from updated cluster
			paralusUtils.BuildResourceFromClusterStruct(a, d)
			break
		}
	}

	paralusUtils.BuildResourceFromClusterStruct(&infrav3.Cluster{}, d)
	return diags

}

// Delete an existing cluster
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	auth := m.(*authprofile.Profile)
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", d.Get("project"), d.Get("name"))

	tflog.Trace(ctx, "Cluster Delete API Request", map[string]interface{}{
		"uri":    uri,
		"method": "DELETE",
	})

	_, err := auth.AuthAndRequest(uri, "DELETE", nil)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to delete cluster %s in project %s", d.Get("name"), d.Get("project"))))
	}

	return diags
}

// Retrieve the YAML files that will be used to setup paralus agents in cluster
func getClusterYAMLs(ctx context.Context, d *schema.ResourceData, auth *authprofile.Profile) diag.Diagnostics {

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s/download", d.Get("project"), d.Get("name"))

	tflog.Trace(ctx, "Cluster YAML GET API Request", map[string]interface{}{
		"uri":    uri,
		"method": "GET",
	})

	resp, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to retrieve K8s YAMLs for cluster %s in project %s", d.Get("name"), d.Get("project"))))
	}

	d.Set("k8s_yamls", resp)
	return nil
}
