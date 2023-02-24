// Group Terraform Resource
package resources

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

// Paralus Resource Group
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource containing paralus group information. Uses the [pctl](https://github.com/paralus/cli) library",
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Group ID in the format \"GROUP_NAME\"",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Group name",
				ForceNew:    true,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Group description.",
				Optional:    true,
			},
			"project_roles": {
				Type:        schema.TypeList,
				Description: "Project namespace roles to attach to the group",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							Description: "Project name",
							Optional:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "Role name",
							Required:    true,
						},
						"namespace": {
							Type:        schema.TypeString,
							Description: "Authorized namespace",
							Optional:    true,
						},
						"group": {
							Type:        schema.TypeString,
							Description: "Authorized group. This will always be the same as the resource group name.",
							Computed:    true,
						},
					},
				},
			},
			"users": {
				Type:        schema.TypeList,
				Description: "User roles attached to group",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of group",
				Optional:    true,
				Default:     "SYSTEM",
			},
		},
	}
}

// Create a specific group
func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	groupId := d.Get("name").(string)

	cfg := m.(*config.Config)
	tflog.Debug(ctx, fmt.Sprintf("resourceGroupCreate provider config used: %s", utils.GetConfigAsMap(cfg)))

	diags := createOrUpdateGroup(ctx, d, "POST", m)

	if diags.HasError() {
		return diags
	}

	d.SetId(groupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	cfg := m.(*config.Config)
	tflog.Debug(ctx, fmt.Sprintf("resourceGroupUpdate provider config used: %s", utils.GetConfigAsMap(cfg)))

	return createOrUpdateGroup(ctx, d, "PUT", m)
}

// Creates a new group or updates an existing one
func createOrUpdateGroup(ctx context.Context, d *schema.ResourceData, requestType string, m interface{}) diag.Diagnostics {

	groupId := d.Get("name").(string)

	auth := m.(*config.Config).GetAppAuthProfile()
	diags := utils.AssertStringNotEmpty("group name", groupId)
	if diags.HasError() {
		return diags
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	tflog.Trace(ctx, fmt.Sprintf("Group %s request", requestType), map[string]interface{}{
		"group": groupId,
	})
	groupStruct := utils.BuildGroupStructFromResource(d)

	// before creating the group, verify that projects in PNR structs exist
	diags = utils.CheckProjectsFromPNRStructExist(groupStruct.Spec.GetProjectNamespaceRoles(), auth)
	if diags.HasError() {
		return diags
	}

	// before creating the group, verify that users in question exist
	diags = utils.CheckUsersExist(groupStruct.Spec.Users, auth)
	if diags.HasError() {
		return diags
	}

	err := utils.ApplyGroup(groupStruct, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err,
			"failed to %s group %s", howFail,
			groupId))
	}

	return resourceGroupRead(ctx, d, m)
}

// Retreive group info
func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	cfg := m.(*config.Config)
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceGroupRead provider config used: %s", utils.GetConfigAsMap(cfg)))

	groupId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("group name", groupId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	groupStruct, err := utils.GetGroupByName(groupId, auth)
	if err == utils.ErrResourceNotExists {
		d.SetId("")
		return diags
	}
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Error retrieving group info for %s", groupId))
	}

	// Update resource information from updated cluster
	utils.BuildResourceFromGroupStruct(groupStruct, d)

	return diags
}

// Import group into TF
func resourceGroupImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	groupId := d.Id()

	cfg := m.(*config.Config)
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceGroupImport provider config used: %s", utils.GetConfigAsMap(cfg)))

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	groupStruct, err := utils.GetGroupByName(groupId, auth)
	// unlike others, fail and stop the import if we fail to get group info
	if err != nil {
		return nil, errors.Wrapf(err, "group %s does not exist", groupId)
	}

	utils.BuildResourceFromGroupStruct(groupStruct, d)

	schemas := make([]*schema.ResourceData, 0)
	schemas = append(schemas, d)
	return schemas, nil

}

// Delete an existing cluster
func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	cfg := m.(*config.Config)
	auth := cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceGroupDelete provider config used: %s", utils.GetConfigAsMap(cfg)))
	groupId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("group name", groupId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Deleting Group info", map[string]interface{}{
		"group": groupId,
	})

	// verify group exists before attempting delete
	_, err := utils.GetGroupByName(groupId, auth)
	if err != nil && err != utils.ErrResourceNotExists {
		return diag.FromErr(errors.Wrapf(err, "failed to retrieve group %s",
			groupId))
	}

	err = utils.DeleteGroup(groupId, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "failed to delete group %s",
			groupId))
	}

	d.SetId("")
	return diags
}
