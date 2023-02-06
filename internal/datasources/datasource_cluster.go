// Cluster Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/cluster"
	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// Paralus DataSource Cluster
func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a paralus cluster's information. Uses the [pctl](https://github.com/paralus/cli) library",
		ReadContext: datasourceClusterRead,
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
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Cluster description",
				Computed:    true,
			},
			"cluster_type": {
				Type:        schema.TypeString,
				Description: "Cluster type. For example, \"imported.\" ",
				Computed:    true,
			},
			"uuid": {
				Type:        schema.TypeString,
				Description: "Cluster UUID",
				Computed:    true,
			},
			"params": {
				Type:        schema.TypeSet,
				Description: "Import parameters",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provision_type": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"IMPORT\"",
							Computed:    true,
						},
						"provision_environment": {
							Type:        schema.TypeString,
							Description: "Provision Environment. For example, \"CLOUD\"",
							Computed:    true,
						},
						"provision_package_type": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"LINUX\"",
							Computed:    true,
						},
						"environment_provider": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"GCP\"",
							Computed:    true,
						},
						"kubernetes_provider": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"EKS\"",
							Computed:    true,
						},
						"state": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"PROVISION\"",
							Computed:    true,
						},
					},
				},
			},
			"project": {
				Type:        schema.TypeString,
				Description: "Project containing cluster",
				Required:    true,
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
				Description: "YAML files used to deploy paralus agent to the cluster stored as a list",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: "Map of lables to include for cluster",
				Computed:    true,
			},
			"annotations": {
				Type:        schema.TypeMap,
				Description: "Map of annotations to include for cluster",
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

// Retreive cluster JSON info
func datasourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterId := d.Get("name").(string)
	projectId := d.Get("project").(string)

	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	if diags.HasError() {
		return diags
	}

	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	if diags.HasError() {
		return diags
	}

	d.SetId(clusterId + ":" + projectId)

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	clusterStruct, err := cluster.GetCluster(clusterId, projectId)

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("error locating cluster %s in project %s",
			clusterId, projectId)))
	}

	utils.BuildResourceFromClusterStruct(clusterStruct, d)

	err = utils.SetBootstrapFileAndRelays(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags

}
