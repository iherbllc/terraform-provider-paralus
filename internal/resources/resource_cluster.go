// Cluster Terraform Resource
package resources

import (
	"context"
	"fmt"
	"strings"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/cluster"
	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/project"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// Paralus Resource Cluster
func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource containing paralus cluster information. Uses the [pctl](https://github.com/paralus/cli) library",
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Cluster ID in the format \"PROJECT_NAME:CLUSTER_NAME\"",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Cluster name",
				Required:    true,
				ForceNew:    true,
			},
			// Must make readonly since paralus API doesn't allow the update
			"description": {
				Type:        schema.TypeString,
				Description: "Cluster description. Paralus API sets it the same as cluster name",
				Computed:    true,
			},
			"cluster_type": {
				Type:        schema.TypeString,
				Description: "Cluster type. For example, \"imported.\" ",
				Optional:    true,
				ForceNew:    true,
			},
			"uuid": {
				Type:        schema.TypeString,
				Description: "Cluster UUID",
				Computed:    true,
			},
			"params": {
				Type:        schema.TypeSet,
				Description: "Import parameters",
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provision_type": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"IMPORT\"",
							Required:    true,
							ForceNew:    true,
						},
						"provision_environment": {
							Type:        schema.TypeString,
							Description: "Provision Environment. For example, \"CLOUD\"",
							Required:    true,
							ForceNew:    true,
						},
						"provision_package_type": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"LINUX\"",
							Optional:    true,
							ForceNew:    true,
						},
						"environment_provider": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"GCP\"",
							Optional:    true,
							ForceNew:    true,
						},
						"kubernetes_provider": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"EKS\"",
							Required:    true,
							ForceNew:    true,
						},
						"state": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"PROVISION\"",
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"project": {
				Type:        schema.TypeString,
				Description: "Project containing cluster",
				Required:    true,
				ForceNew:    true,
			},
			// Will only ever be updated by provider
			"bootstrap_files_combined": {
				Type:        schema.TypeString,
				Description: "YAML files used to deploy paralus agent to the cluster stored as a single massive file",
				Computed:    true,
			},
			// Will only ever be updated by provider
			"bootstrap_files": {
				Type:        schema.TypeList,
				Description: "YAML files used to deploy paralus agent to the cluster stored as a list of files",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Can be passed in or updated by provider
			// A newly created cluster will have it's labels added to by paralus
			"labels": {
				Type:        schema.TypeMap,
				Description: "Map of lables to include for cluster",
				Optional:    true,
				Computed:    true,
			},
			// Can be passed in or updated by provider
			// A newly created cluster will have it's annotations added to by paralus
			"annotations": {
				Type:        schema.TypeMap,
				Description: "Map of annotations to include for cluster",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Create a new cluster in Paralus
func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(m.(*config.Config))))

	diags := append(createOrUpdateCluster(ctx, d, "POST"), setBootstrapFile(ctx, d)...)

	d.SetId(projectId + ":" + clusterId)

	return diags
}

// Updating existing cluster
func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(m.(*config.Config))))
	return createOrUpdateCluster(ctx, d, "PUT")
}

// Creates a new cluster or updates an existing one
func createOrUpdateCluster(ctx context.Context, d *schema.ResourceData, requestType string) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	tflog.Trace(ctx, fmt.Sprintf("Checking for project %s existance", projectId))

	projectStruct, err := project.GetProjectByName(projectId)
	if projectStruct == nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Project %s does not exist", projectId)))
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	clusterStruct := paralusUtils.BuildClusterStructFromResource(d)

	tflog.Trace(ctx, fmt.Sprintf("Cluster %s request", requestType), map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	if requestType == "POST" {
		err := cluster.CreateCluster(clusterStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
					clusterId, projectId)))
		}
	} else if requestType == "PUT" {
		err := cluster.UpdateCluster(clusterStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
					clusterId, projectId)))
		}
	} else {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Unknown request type %s", requestType)))
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	importedStruct, err := cluster.GetCluster(clusterId, projectId)

	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("Failed to %s cluster %s in project %s", howFail,
				clusterId, projectId)))
	}

	// Update resource information from created/updated cluster
	paralusUtils.BuildResourceFromClusterStruct(importedStruct, d)

	return diags
}

// Retreive cluster info
func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(m.(*config.Config))))

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	_, err := cluster.GetCluster(clusterId, projectId)

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Cluster %s does not exist in project %s",
			clusterId, projectId)))
	}

	if d.Id() == "" {
		d.SetId(projectId + ":" + clusterId)
	}

	return diags

}

// Import cluster info into TF
func resourceClusterImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(m.(*config.Config))))

	clusterProjectId := strings.Split(d.Id(), ":")

	if len(clusterProjectId) != 2 {
		d.SetId("")
		return nil, errors.Wrap(nil, fmt.Sprintf("Unable to import. ID must be in format PROJECT_NAME:CLUSTER_NAME. Got %s", d.Id()))
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"project": clusterProjectId[0],
		"cluster": clusterProjectId[1],
	})

	clusterStruct, err := cluster.GetCluster(clusterProjectId[1], clusterProjectId[0])

	if err != nil {
		d.SetId("")
		return nil, errors.Wrap(err, fmt.Sprintf("Cluster %s does not exist in project %s",
			clusterProjectId[1], clusterProjectId[0]))
	}

	paralusUtils.BuildResourceFromClusterStruct(clusterStruct, d)
	setBootstrapFile(ctx, d)

	schemas := make([]*schema.ResourceData, 0)
	schemas = append(schemas, d)
	return schemas, nil

}

// Delete an existing cluster
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", paralusUtils.GetConfigAsMap(m.(*config.Config))))

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	tflog.Trace(ctx, "Deleting cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	// Assume if uuid is not set, then the cluster was not created
	// So skip the delete
	// This is to avoid the situation where the failure is due to an invalid endpoint, which would
	// fail the acct test as well
	if d.Get("uuid") == "" {
		d.SetId("")
		return diags
	}

	err := cluster.DeleteCluster(clusterId, projectId)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed to delete cluster %s in project %s",
			clusterId, projectId)))
	}

	d.SetId("")
	return diags
}

// Retrieve the YAML files that will be used to setup paralus agents in cluster and assign it to the schema
func setBootstrapFile(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	tflog.Trace(ctx, "Retrieving Bootstrap File", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	// Make sure cluster exists before we attempt to get the bootstrap file
	clusterStruct, _ := cluster.GetCluster(clusterId, projectId)
	if clusterStruct == nil {
		d.SetId("")
		return nil
	}

	bootstrapFile, err := cluster.GetBootstrapFile(clusterId, projectId)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error retrieving bootstrap file for cluster %s in project %s",
			clusterId, projectId)))
	}

	d.Set("bootstrap_files_combined", bootstrapFile)
	d.Set("bootstrap_files", paralusUtils.SplitSingleYAMLIntoList(bootstrapFile))
	return nil
}
