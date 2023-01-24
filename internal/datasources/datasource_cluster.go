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
		ReadContext: datasourceClusterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
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

// Retreive cluster JSON info
func datasourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(d.Get("name").(string) + d.Get("project").(string) + "ds")

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": d.Get("name").(string),
		"project": d.Get("project").(string),
	})

	cluster, err := cluster.GetCluster(d.Get("name").(string), d.Get("project").(string))

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error locating cluster %s in project %s",
			d.Get("name").(string), d.Get("project").(string))))
	}

	paralusUtils.BuildResourceFromClusterStruct(cluster, d)
	return diags

}
