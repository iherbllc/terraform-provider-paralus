// Cluster Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/cluster"

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
			"bootstrap_file": {
				Type:        schema.TypeString,
				Description: "YAML files used to deploy paralus agent to the cluster",
				Computed:    true,
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
		},
	}
}

// Retreive cluster JSON info
func datasourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterId := d.Get("name").(string)
	projectId := d.Get("project").(string)

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	clusterStruct, err := cluster.GetCluster(clusterId, projectId)

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error locating cluster %s in project %s",
			clusterId, projectId)))
	}

	paralusUtils.BuildResourceFromClusterStruct(clusterStruct, d)

	bootstrapFile, err := cluster.GetBootstrapFile(clusterId, projectId)

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error retrieving bootstrap file for cluster %s in project %s",
			clusterId, projectId)))
	}

	d.Set("bootstrap_file", bootstrapFile)

	d.SetId(clusterId + ":" + projectId)

	return diags

}
