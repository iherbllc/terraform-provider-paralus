// Group Terraform Resource
package resources

import (
	"context"
	"fmt"

	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = (*RsGoup)(nil)

func ResourceGroup() resource.Resource {
	return &RsGoup{}
}

type RsGoup struct {
	cfg *config.Config
}

// With the resource.Resource implementation
func (r *RsGoup) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Paralus Resource Group
func (r RsGoup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource containing paralus group information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Group ID in the format \"GROUP_NAME\"",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				DeprecationMessage: "id is no longer a required valie for providers and will eventually be removed. Use \"name\" instead.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Group name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Group description.",
				Optional:            true,
			},
			"users": schema.ListAttribute{
				MarkdownDescription: "User roles attached to group",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of group",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("SYSTEM"),
			},
		},
		Blocks: map[string]schema.Block{
			"project_roles": schema.ListNestedBlock{
				MarkdownDescription: "Project namespace roles to attach to the group",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"project": schema.StringAttribute{
							MarkdownDescription: "Project name",
							Optional:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Role name",
							Required:            true,
						},
						"namespace": schema.StringAttribute{
							MarkdownDescription: "Authorized namespace",
							Optional:            true,
						},
						"group": schema.StringAttribute{
							MarkdownDescription: "Authorized group. This will always be the same as the resource group name.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *RsGoup) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.cfg = cfg
}

// Create a specific group
func (r *RsGoup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Group
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Create provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	diags = createOrUpdateGroup(ctx, data, "POST", r.cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r RsGoup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Group
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Update provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	diags = createOrUpdateGroup(ctx, data, "PUT", r.cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Creates a new group or updates an existing one
func createOrUpdateGroup(ctx context.Context, data *structs.Group, requestType string, cfg *config.Config) diag.Diagnostics {

	var diags diag.Diagnostics
	groupId := data.Name.ValueString()

	auth := cfg.GetAppAuthProfile()
	diags = utils.AssertStringNotEmpty("group name", groupId)
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
	groupStruct, diags := utils.BuildGroupStructFromResource(ctx, data)
	if diags.HasError() {
		return diags
	}

	// before creating the group, verify that projects in PNR structs exist
	diags = utils.CheckProjectsFromPNRStructExist(ctx, groupStruct.Spec.GetProjectNamespaceRoles(), auth)
	if diags.HasError() {
		return diags
	}

	// need to make sure the combination of namespace, group, project, and role are all unique per entry
	diags = utils.AssertUniquePRNStruct(groupStruct.Spec.GetProjectNamespaceRoles())
	if diags.HasError() {
		return diags
	}

	// before creating the group, verify that users in question exist
	diags = utils.CheckUsersExist(ctx, groupStruct.Spec.Users, auth)
	if diags.HasError() {
		return diags
	}

	err := utils.ApplyGroup(ctx, groupStruct, auth)
	if err != nil {
		diags.AddError(fmt.Sprintf(
			"failed to %s group %s", howFail,
			groupId), err.Error())
		return diags
	}

	// Update resource information from updated group
	diags = utils.BuildResourceFromGroupStruct(ctx, groupStruct, data)
	return diags
}

// Retreive group info
func (r RsGoup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Group
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	groupId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("group name", groupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	groupStruct, err := utils.GetGroupByName(ctx, groupId, auth)
	if err == utils.ErrResourceNotExists {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving group info for %s", groupId), err.Error())
		return
	}

	// Update resource information from updated group
	diags = utils.BuildResourceFromGroupStruct(ctx, groupStruct, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Import group into TF
func (r *RsGoup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceGroupImport provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	groupId := req.ID
	if groupId == "" || groupId == "id-attribute-not-set" {
		resp.Diagnostics.AddError("Must specify a group name when importing", "")
		return
	}

	tflog.Trace(ctx, "Retrieving group info", map[string]interface{}{
		"group": groupId,
	})

	groupStruct, err := utils.GetGroupByName(ctx, groupId, auth)
	// unlike others, fail and stop the import if we fail to get group info
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("group %s does not exist", req.ID),
		)
		return
	}

	var data structs.Group
	diags := utils.BuildResourceFromGroupStruct(ctx, groupStruct, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Delete an existing group
func (r RsGoup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Group
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Delete provider config used: %s", utils.GetConfigAsMap(r.cfg)))
	groupId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("group name", groupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleting Group info", map[string]interface{}{
		"group": groupId,
	})

	// verify group exists before attempting delete
	_, err := utils.GetGroupByName(ctx, groupId, auth)
	if err != nil && err != utils.ErrResourceNotExists {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve group %s",
			groupId), err.Error())
		return
	}

	err = utils.DeleteGroup(ctx, groupId, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete group %s",
			groupId), err.Error())
	}
}
