// KubeConfig Terraform DataSource
package datasources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Paralus DataSource KubeConfig
func DataSourceKubeConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a user's kubeconfig information. Uses the [pctl](https://github.com/paralus/cli) library",
		ReadContext: datasourceKubeConfigRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "User's name",
				Required:    true,
			},
			"namespace": {
				Type:        schema.TypeString,
				Description: "Namespace to set as the default for the kubeconfig",
				Optional:    true,
			},
			"cluster": {
				Type:        schema.TypeString,
				Description: "Cluster to get certificate information for",
				Optional:    true,
			},
			"cluster_info": {
				Type:        schema.TypeList,
				Description: "KubeConfig cluster information",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_authority_data": {
							Type:        schema.TypeString,
							Description: "Certificate authority data for cluster",
							Computed:    true,
							Sensitive:   true,
						},
						"server": {
							Type:        schema.TypeString,
							Description: "URL to server",
							Computed:    true,
						},
					},
				},
			},
			"client_certificate_data": {
				Type:        schema.TypeString,
				Description: "Client certificate data",
				Computed:    true,
				Sensitive:   true,
			},
			"client_key_data": {
				Type:        schema.TypeString,
				Description: "Client key data",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Retreive KubeConfig JSON info
func datasourceKubeConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	userName := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("name", userName)
	if diags.HasError() {
		return diags
	}

	cluster := d.Get("cluster").(string)
	namespace := d.Get("namespace").(string)

	cfg := m.(*config.Config)
	auth := cfg.GetAppAuthProfile()
	userInfo, err := utils.GetUserByName(ctx, userName, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "error locating user info: %s", userName))
	}

	userID := userInfo.Metadata.Id

	tflog.Trace(ctx, "Retrieving KubeConfig info", map[string]interface{}{
		"name":      userName,
		"id":        userID,
		"cluster":   cluster,
		"namespace": namespace,
	})

	tflog.Debug(ctx, fmt.Sprintf("datasourceKubeConfigRead provider config used: %s", utils.GetConfigAsMap(cfg)))

	kubeConfig, err := utils.GetKubeConfig(ctx, userID, namespace, cluster, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "error locating kubeconfig for user %s. Make sure the kubeconfig has been generated manually through the UI for the first time.", userName))
	}

	utils.BuildKubeConfigStruct(ctx, d, kubeConfig)

	d.SetId(userName)

	return diags

}
