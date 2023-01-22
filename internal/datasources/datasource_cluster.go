package datasources

import (
	"context"
	"fmt"

	paralusUtils "github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/authprofile"

	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"

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

// Retreive cluster JSON info
func datasourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
