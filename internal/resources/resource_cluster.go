package resources

import (
	"context"
	"fmt"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/cluster"
	"github.com/paralus/cli/pkg/project"

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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
			"params": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
					},
				},
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bootstrap_file": {
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
		},
	}
}

// Import an existing K8S cluster into a designated project
func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	project := d.Get("project").(string)
	cluster := d.Get("name").(string)

	diags := append(createOrUpdateCluster(ctx, d, "POST"), getClusterYAMLs(ctx, d)...)

	d.SetId(project + ":" + cluster)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateCluster(ctx, d, "PUT")
}

// Creates a new cluster or updates an existing one
func createOrUpdateCluster(ctx context.Context, d *schema.ResourceData, requestType string) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Trace(ctx, fmt.Sprintf("Checking for project %s existance", d.Get("project")))

	projectStruct, err := project.GetProjectByName(d.Get("project").(string))
	if projectStruct == nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Project %s does not exist", d.Get("project"))))
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	clusterStruct := paralusUtils.BuildClusterStructFromResource(d)

	tflog.Trace(ctx, fmt.Sprintf("Cluster %s request", requestType), map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	if requestType == "POST" {
		err := cluster.CreateCluster(clusterStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
					d.Get("name"), d.Get("project"))))
		}
	} else if requestType == "PUT" {
		err := cluster.UpdateCluster(clusterStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
					d.Get("name"), d.Get("project"))))
		}
	} else {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Unknown request type %s", requestType)))
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	_, err = cluster.GetCluster(d.Get("name").(string), d.Get("project").(string))

	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
				d.Get("name"), d.Get("project"))))
	}

	// Update resource information from updated cluster
	paralusUtils.BuildResourceFromClusterStruct(clusterStruct, d)

	return diags
}

// Retreive cluster JSON info
func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	_, err := cluster.GetCluster(d.Get("name").(string), d.Get("project").(string))

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Cluster %s does not exist in project %s",
			d.Get("name").(string), d.Get("project").(string))))
	}

	return diags

}

// Delete an existing cluster
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Trace(ctx, "Deleting cluster info", map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	// Make sure cluster exists before we attempt to delete it
	clusterStruct, _ := cluster.GetCluster(d.Get("name").(string), d.Get("project").(string))
	if clusterStruct == nil {
		d.SetId("")
		return diags
	}

	err := cluster.DeleteCluster(d.Get("name").(string), d.Get("project").(string))

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed to delete cluster %s in project %s",
			d.Get("name").(string), d.Get("project").(string))))
	}

	d.SetId("")
	return diags
}

// Retrieve the YAML files that will be used to setup paralus agents in cluster
func getClusterYAMLs(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {

	tflog.Trace(ctx, "Retrieving Bootstrap File", map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	// Make sure cluster exists before we attempt to get the bootstrap file
	clusterStruct, _ := cluster.GetCluster(d.Get("project").(string), d.Get("name").(string))
	if clusterStruct == nil {
		return nil
	}

	bootstrapFile, err := cluster.GetBootstrapFile(d.Get("project").(string), d.Get("name").(string))

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error retrieving bootstrap file for cluster %s in project %s",
			d.Get("name").(string), d.Get("project").(string))))
	}

	d.Set("bootstrap_file", bootstrapFile)
	return nil
}
