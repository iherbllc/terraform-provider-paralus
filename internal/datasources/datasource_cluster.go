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

// / Paralus DataSource Cluster
func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a paralus cluster's information. Uses the [pctl|https://github.com/paralus/cli] library",
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
				Optional:    true,
			},
			"cluster_type": {
				Type:        schema.TypeString,
				Description: "Cluster type. For example, \"imported.\" ",
				Optional:    true,
			},
			"params": {
				Type:        schema.TypeSet,
				Description: "Import parameters",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provision_type": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"IMPORT\"",
							Required:    true,
						},
						"provision_environment": {
							Type:        schema.TypeString,
							Description: "Provision Environment. For example, \"CLOUD\"",
							Required:    true,
						},
						"provision_package_type": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"LINUX\"",
							Optional:    true,
						},
						"environment_provider": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"GCP\"",
							Optional:    true,
						},
						"kubernetes_provider": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"EKS\"",
							Required:    true,
						},
						"state": {
							Type:        schema.TypeString,
							Description: "Provision Type. For example, \"PROVISION\"",
							Required:    true,
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
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: "Map of lables to include for cluster",
				Optional:    true,
			},
			"annotations": {
				Type:        schema.TypeMap,
				Description: "Map of annotations to include for cluster",
				Optional:    true,
			},
		},
	}
}

// Retreive cluster JSON info
func datasourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	clusterStruct, err := cluster.GetCluster(d.Get("name").(string), d.Get("project").(string))

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error locating cluster %s in project %s",
			d.Get("name").(string), d.Get("project").(string))))
	}

	paralusUtils.BuildResourceFromClusterStruct(clusterStruct, d)

	bootstrapFile, err := cluster.GetBootstrapFile(d.Get("project").(string), d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error retrieving bootstrap file for cluster %s in project %s",
			d.Get("name").(string), d.Get("project").(string))))
	}

	d.Set("bootstrap_file", bootstrapFile)

	d.SetId(d.Get("name").(string) + ":" + d.Get("project").(string))

	return diags

}
