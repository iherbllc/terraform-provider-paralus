// Project Terraform Resource
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = (*RsProject)(nil)

func ResourceProject() resource.Resource {
	return &RsProject{}
}

type RsProject struct {
	cfg *config.Config
}

// With the resource.Resource implementation
func (r *RsProject) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Paralus Resource Group
// Paralus Resource Cluster
func (r RsProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource containing paralus project information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project ID in the format \"PROJECT_NAME\"",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Project description.",
				Optional:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Project UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_roles": schema.ListNestedAttribute{
				MarkdownDescription: "Project roles attached to project, containing group or namespace",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project": schema.StringAttribute{
							MarkdownDescription: "Project name. This will always be the same as the resource project name.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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
							MarkdownDescription: "Authorized group",
							Required:            true,
						},
					},
				},
			},
			"user_roles": schema.ListNestedAttribute{
				MarkdownDescription: "User roles attached to project",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.StringAttribute{
							MarkdownDescription: "Authorized user",
							Required:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Authorized role",
							Required:            true,
						},
						"namespace": schema.StringAttribute{
							MarkdownDescription: "Authorized namespace",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *RsProject) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Import an existing K8S cluster into a designated project
func (r *RsProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Project
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create provider config used: %s", utils.GetConfigAsMap(r.cfg)))
	diags = createOrUpdateProject(ctx, data, "POST", r.cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r RsProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Project
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectUpdate provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	diags = createOrUpdateProject(ctx, data, "PUT", r.cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Creates a new project or updates an existing one
func createOrUpdateProject(ctx context.Context, data *structs.Project, requestType string, cfg *config.Config) diag.Diagnostics {

	var diags diag.Diagnostics
	projectId := data.Name.ValueString()

	auth := cfg.GetAppAuthProfile()
	diags = utils.AssertStringNotEmpty("project name", projectId)
	if diags.HasError() {
		return diags
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	tflog.Trace(ctx, fmt.Sprintf("Project %s request", requestType), map[string]interface{}{
		"project": projectId,
	})

	projectStruct, diags := utils.BuildProjectStructFromResource(ctx, data)
	if diags.HasError() {
		return diags
	}

	// before creating the project, verify that requested group exists
	diags = utils.CheckGroupsFromPNRStructExist(ctx, projectStruct.Spec.GetProjectNamespaceRoles(), auth)
	if diags.HasError() {
		return diags
	}

	// before creating the project, verify that users in question exist
	diags = utils.CheckUserRoleUsersExist(ctx, projectStruct.Spec.GetUserRoles(), auth)
	if diags.HasError() {
		return diags
	}

	// Required due to limitation with paralus
	// See issue: https://github.com/paralus/paralus/issues/136
	// Remove once paralus supports this feature.
	diags = utils.AssertUniqueRoles(projectStruct.Spec.ProjectNamespaceRoles)
	if diags.HasError() {
		return diags
	}

	err := utils.ApplyProject(ctx, projectStruct, auth)
	if err != nil {
		diags.AddError(fmt.Sprintf(
			"failed to %s project %s", howFail,
			projectId), err.Error())
	}

	return diags
}

// Retreive project info
func (r RsProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Project
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	projectId := data.Name.ValueString()
	diags = utils.AssertStringNotEmpty("project name", projectId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	projectStruct, err := utils.GetProjectByName(ctx, projectId, auth)
	if err == utils.ErrResourceNotExists {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving info for project %s", projectId), err.Error())
	}

	// Update resource information from updated cluster
	diags = utils.BuildResourceFromProjectStruct(ctx, projectStruct, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Import project into TF
func (r *RsProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Project
	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Import provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	projectId := req.ID
	if projectId == "" {
		resp.Diagnostics.AddError("Must specify a project name", "")
		return
	}

	tflog.Trace(ctx, "Retrieving project info", map[string]interface{}{
		"project": projectId,
	})

	projectStruct, err := utils.GetProjectByName(ctx, projectId, auth)
	// unlike others, fail and stop the import if we fail to get project info
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("project %s does not exist", req.ID),
		)
		return
	}

	diags := utils.BuildResourceFromProjectStruct(ctx, projectStruct, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Delete an existing cluster
func (r RsProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Project
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("resourceProjectDelete provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	projectId := data.Name.ValueString()
	diags = utils.AssertStringNotEmpty("project name", projectId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleting Project info", map[string]interface{}{
		"project": projectId,
	})

	// verify project exists before attempting delete
	_, err := utils.GetProjectByName(ctx, projectId, auth)
	if err != nil && err != utils.ErrResourceNotExists {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve project %s",
			projectId), err.Error())
		return
	}

	err = utils.DeleteProject(ctx, projectId, auth)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete project %s",
			projectId), err.Error())
		return
	}
}
