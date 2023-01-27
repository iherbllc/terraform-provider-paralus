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

// Paralus DataSource Bootstrap File
func DataSourceBootstrapFile() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the bootstrap file generated after a cluster is imported. Uses the [pctl](https://github.com/paralus/cli) library",
		ReadContext: datasourceBootstrapFileRead,
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
		},
	}
}

// Retreive cluster bootstrap file
func datasourceBootstrapFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterId := d.Get("name").(string)
	projectId := d.Get("project").(string)

	tflog.Trace(ctx, "Retrieving bootstrap info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	_, err := cluster.GetCluster(clusterId, projectId)

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error locating cluster %s in project %s",
			clusterId, projectId)))
	}

	bootstrapFile, err := cluster.GetBootstrapFile(clusterId, projectId)

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error retrieving bootstrap file for cluster %s in project %s",
			clusterId, projectId)))
	}

	d.Set("bootstrap_files_combined", bootstrapFile)
	d.Set("bootstrap_files", paralusUtils.SplitSingleYAMLIntoList(bootstrapFile))

	d.SetId(clusterId + ":" + projectId)

	return diags

}
