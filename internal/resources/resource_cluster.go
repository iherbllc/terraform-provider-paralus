// Cluster Terraform Resource
package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

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
			"relays": {
				Type:        schema.TypeString,
				Description: "Relays information",
				Computed:    true,
			},
		},
	}
}

// Create a new cluster in Paralus
func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	diags := createOrUpdateCluster(ctx, d, "POST")
	if diags.HasError() {
		return diags
	}

	d.SetId(d.Get("project").(string) + ":" + d.Get("name").(string))

	return nil
}

// Updating existing cluster
func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	return createOrUpdateCluster(ctx, d, "PUT")
}

// Creates a new cluster or updates an existing one
func createOrUpdateCluster(ctx context.Context, d *schema.ResourceData, requestType string) diag.Diagnostics {

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	diags := utils.AssertStringNotEmpty("cluster project", projectId)
	if diags.HasError() {
		return diags
	}

	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, fmt.Sprintf("Checking for project %s existance", projectId))

	projectStruct, err := project.GetProjectByName(projectId)
	if projectStruct == nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("project %s does not exist", projectId)))
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	clusterStruct := utils.BuildClusterStructFromResource(d)

	tflog.Trace(ctx, fmt.Sprintf("Cluster %s request", requestType), map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	if requestType == "POST" {
		// due to error swallowing, have to make sure the cluster doesn't exist before
		// attempting to create it.
		lookupStruct, _ := utils.GetCluster(clusterId, projectId)
		if lookupStruct != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("cluster %s already exists", clusterId)))
		}

		err := utils.CreateCluster(clusterStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("failed to %s cluster %s in project %s", howFail,
					clusterId, projectId)))
		}
	} else if requestType == "PUT" {
		err := utils.UpdateCluster(clusterStruct)
		if err != nil {
			return diag.FromErr(errors.Wrap(err,
				fmt.Sprintf("failed to %s cluster %s in project %s", howFail,
					clusterId, projectId)))
		}
	} else {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("unknown request type %s", requestType)))
	}

	return resourceClusterRead(ctx, d, nil)
}

// Retreive cluster info
func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	if diags.HasError() {
		return diags
	}

	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	clusterStruct, err := utils.GetCluster(clusterId, projectId)

	tflog.Trace(ctx, fmt.Sprintf("ClusterStruct from GetCluster: %v", clusterStruct))
	tflog.Trace(ctx, fmt.Sprintf("Error from GetCluster: %s", err))

	if clusterStruct == nil {
		d.SetId("")
		return diags
	}

	// Update resource information from created/updated cluster
	utils.BuildResourceFromClusterStruct(clusterStruct, d)

	err = utils.SetBootstrapFileAndRelays(ctx, d)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "called from resourceClusterRead"))
	}

	return diags
}

// Import cluster info into TF
func resourceClusterImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	clusterProjectId := strings.Split(d.Id(), ":")

	if len(clusterProjectId) != 2 {
		d.SetId("")
		return nil, errors.Wrap(nil, fmt.Sprintf("unable to import. ID must be in format PROJECT_NAME:CLUSTER_NAME. Got %s", d.Id()))
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"project": clusterProjectId[0],
		"cluster": clusterProjectId[1],
	})

	clusterStruct, err := utils.GetCluster(clusterProjectId[1], clusterProjectId[0])

	if err != nil {
		d.SetId("")
		// unlike others, we want to throw an error if the cluster does not exist so we can fail the import
		return nil, errors.Wrap(err, fmt.Sprintf("cluster %s does not exist in project %s",
			clusterProjectId[1], clusterProjectId[0]))
	}

	utils.BuildResourceFromClusterStruct(clusterStruct, d)
	err = utils.SetBootstrapFileAndRelays(ctx, d)
	if err != nil {
		d.SetId("")
		return nil, errors.Wrap(err, "called from resourceClusterImport")
	}

	schemas := make([]*schema.ResourceData, 0)
	schemas = append(schemas, d)
	return schemas, nil

}

// Delete an existing cluster
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	projectId := d.Get("project").(string)
	clusterId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	if diags.HasError() {
		return diags
	}

	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Deleting cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	clusterStruct, _ := utils.GetCluster(clusterId, projectId)
	if clusterStruct != nil {
		err := utils.DeleteCluster(clusterId, projectId)

		if err != nil {
			return diag.FromErr(errors.Wrap(err, fmt.Sprintf("failed to delete cluster %s in project %s",
				clusterId, projectId)))
		}
	}

	d.SetId("")
	return diags
}
