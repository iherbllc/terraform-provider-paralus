// Group Terraform Resource
package resources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/config"
	"github.com/paralus/cli/pkg/group"

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
							Description: "Authorized group",
							Required:    true,
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
				Computed:    true,
			},
		},
	}
}

// Create a specific group
func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	groupId := d.Get("name").(string)

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	diags := createOrUpdateGroup(ctx, d, "POST")

	if diags.HasError() {
		return diags
	}

	d.SetId(groupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))
	return createOrUpdateGroup(ctx, d, "PUT")
}

// Creates a new group or updates an existing one
func createOrUpdateGroup(ctx context.Context, d *schema.ResourceData, requestType string) diag.Diagnostics {

	groupId := d.Get("name").(string)

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

	err := group.ApplyGroup(groupStruct)
	if err != nil {
		return diag.FromErr(errors.Wrap(err,
			fmt.Sprintf("failed to %s group %s", howFail,
				groupId)))
	}

	return resourceGroupRead(ctx, d, nil)
}

// Retreive group info
func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	groupId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("group name", groupId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	groupStruct, err := group.GetGroupByName(groupId)
	if groupStruct == nil {
		// error should be "no rows in result set" but add it to TRACE in case it isn't.
		tflog.Trace(ctx, fmt.Sprintf("Error retrieving group info: %s", err))
		d.SetId("")
		return diags
	}

	// Update resource information from updated cluster
	utils.BuildResourceFromGroupStruct(groupStruct, d)

	return diags
}

// Import group into TF
func resourceGroupImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	groupId := d.Id()

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	groupStruct, err := group.GetGroupByName(groupId)
	// unlike others, fail and stop the import if we fail to get group info
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("group %s does not exist", groupId))
	}

	utils.BuildResourceFromGroupStruct(groupStruct, d)

	schemas := make([]*schema.ResourceData, 0)
	schemas = append(schemas, d)
	return schemas, nil

}

// Delete an existing cluster
func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Provider Config Used: %s", utils.GetConfigAsMap(config.GetConfig())))

	groupId := d.Get("name").(string)

	diags = utils.AssertStringNotEmpty("group name", groupId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, "Deleting Group info", map[string]interface{}{
		"group": groupId,
	})

	// verify group exists before attempting delete
	groupStruct, _ := group.GetGroupByName(groupId)
	if groupStruct != nil {

		err := group.DeleteGroup(groupId)
		if err != nil {
			return diag.FromErr(errors.Wrap(err, fmt.Sprintf("failed to delete group %s",
				groupId)))
		}
	}

	d.SetId("")
	return diags
}
