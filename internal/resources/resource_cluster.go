// Cluster Terraform Resource
package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = (*RsCluster)(nil)

func ResourceCluster() resource.Resource {
	return &RsCluster{}
}

type RsCluster struct {
	cfg *config.Config
}

// With the resource.Resource implementation
func (r *RsCluster) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Paralus Resource Cluster
func (r RsCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource containing paralus cluster information. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster ID in the format \"PROJECT_NAME:CLUSTER_NAME\"",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Cluster name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Must make readonly since paralus API doesn't allow the update
			"description": schema.StringAttribute{
				MarkdownDescription: "Cluster description. Paralus API sets it the same as cluster name",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_type": schema.StringAttribute{
				MarkdownDescription: "Type of cluster being created. For example, \"imported\"",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Cluster UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"params": schema.MapNestedAttribute{
				MarkdownDescription: "Import parameters",
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"provision_type": schema.StringAttribute{
							MarkdownDescription: "Provision Type. For example, \"IMPORT\"",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"provision_environment": schema.StringAttribute{
							MarkdownDescription: "Provision Environment. For example, \"CLOUD\"",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"provision_package_type": schema.StringAttribute{
							MarkdownDescription: "Provision Type. For example, \"LINUX\"",
							Optional:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"environment_provider": schema.StringAttribute{
							MarkdownDescription: "Provision Type. For example, \"GCP\"",
							Optional:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"kubernetes_provider": schema.StringAttribute{
							MarkdownDescription: "Provision Type. For example, \"EKS\"",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "Provision Type. For example, \"PROVISION\"",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"project": schema.StringAttribute{
				MarkdownDescription: "Project containing cluster",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Will only ever be updated by provider
			"bootstrap_files_combined": schema.StringAttribute{
				MarkdownDescription: "YAML files used to deploy paralus agent to the cluster stored as a single massive file",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// Will only ever be updated by provider
			"bootstrap_files": schema.ListAttribute{
				MarkdownDescription: "YAML files used to deploy paralus agent to the cluster stored as a list of files",
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			// Can be passed in or updated by provider
			// A newly created cluster will have it's labels added to by paralus
			"labels": schema.MapAttribute{
				MarkdownDescription: "Map of lables to include for cluster",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			// Can be passed in or updated by provider
			// A newly created cluster will have it's annotations added to by paralus
			"annotations": schema.MapAttribute{
				MarkdownDescription: "Map of annotations to include for cluster",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"relays": schema.StringAttribute{
				MarkdownDescription: "Relays information",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RsCluster) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create a new cluster in Paralus
func (r *RsCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Cluster
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Create provider Config Used: %s", utils.GetConfigAsMap(r.cfg)))

	diags = createOrUpdateCluster(ctx, data, "POST", r.cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Updating existing cluster
func (r RsCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Cluster
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Update provider Config Used: %s", utils.GetConfigAsMap(r.cfg)))

	diags = createOrUpdateCluster(ctx, data, "POST", r.cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Creates a new cluster or updates an existing one
func createOrUpdateCluster(ctx context.Context, data *structs.Cluster, requestType string, cfg *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId := data.Project.ValueString()
	clusterId := data.Name.ValueString()

	auth := cfg.GetAppAuthProfile()
	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	if diags.HasError() {
		return diags
	}
	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, fmt.Sprintf("Checking for project %s existance", projectId))

	projectStruct, err := utils.GetProjectByName(ctx, projectId, auth)
	if projectStruct == nil {
		diags.AddError(fmt.Sprintf("project %s does not exist", projectId), err.Error())
		return diags
	}

	howFail := "create"
	if requestType == "PUT" {
		howFail = "update"
	}

	clusterStruct, diags := utils.BuildClusterStructFromResource(ctx, data)
	if diags.HasError() {
		return diags
	}

	tflog.Trace(ctx, fmt.Sprintf("Cluster %s request", requestType), map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	if requestType == "POST" {
		lookupStruct, err := utils.GetCluster(ctx, clusterId, projectId, auth)
		if lookupStruct != nil {
			diags.AddError(fmt.Sprintf("cluster %s in project %s already exists", clusterId, projectId), "")
			return diags
		}
		if err != nil && err != utils.ErrResourceNotExists {
			diags.AddError(fmt.Sprintf("failed to get cluster %s in project %s", clusterId, projectId), err.Error())
			return diags
		}

		err = utils.CreateCluster(ctx, clusterStruct, auth)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to %s cluster %s in project %s", howFail, clusterId, projectId), err.Error())
		}
	} else if requestType == "PUT" {
		err := utils.UpdateCluster(ctx, clusterStruct, auth)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to %s cluster %s in project %s", howFail, clusterId, projectId), err.Error())
		}
	} else {
		diags.AddError(fmt.Sprintf("unknown request type %s", requestType), "")
	}

	return diags
}

// Retreive cluster info
func (r RsCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Cluster
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Read provider Config Used: %s", utils.GetConfigAsMap(r.cfg)))

	projectId := data.Project.ValueString()
	clusterId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	resp.Diagnostics.Append(diags...)
	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	clusterStruct, err := utils.GetCluster(ctx, clusterId, projectId, auth)

	tflog.Trace(ctx, fmt.Sprintf("ClusterStruct from GetCluster: %v", clusterStruct))
	tflog.Trace(ctx, fmt.Sprintf("Error from GetCluster: %s", err))

	if err == utils.ErrResourceNotExists {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error retrieving info for cluster %s in project %s", clusterId, projectId), err.Error())
		return
	}

	// Update resource information from created/updated cluster
	diags = utils.BuildResourceFromClusterStruct(ctx, clusterStruct, data, auth)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Import cluster info into TF
func (r *RsCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Cluster
	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Import provider Config Used: %s", utils.GetConfigAsMap(r.cfg)))

	clusterProjectId := strings.Split(req.ID, ":")

	if len(clusterProjectId) != 2 || clusterProjectId[0] == "" || clusterProjectId[1] == "" {
		resp.Diagnostics.AddError("Unexpected Import Identifier", fmt.Sprintf("unable to import. ID must be in format PROJECT_NAME:CLUSTER_NAME. Got %s", req.ID))
		return
	}

	tflog.Trace(ctx, "Retrieving cluster info", map[string]interface{}{
		"project": clusterProjectId[0],
		"cluster": clusterProjectId[1],
	})

	clusterStruct, err := utils.GetCluster(ctx, clusterProjectId[1], clusterProjectId[0], auth)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cluster %s does not exist in project %s",
			clusterProjectId[1], clusterProjectId[0]), "")
		return
	}

	diags := utils.BuildResourceFromClusterStruct(ctx, clusterStruct, data, auth)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

// Delete an existing cluster
func (r RsCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Prevent panic if the provider has not been configured.
	if r.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.Cluster
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := r.cfg.GetAppAuthProfile()
	tflog.Debug(ctx, fmt.Sprintf("Delete provider config used: %s", utils.GetConfigAsMap(r.cfg)))

	projectId := data.Project.ValueString()
	clusterId := data.Name.ValueString()

	diags = utils.AssertStringNotEmpty("cluster project", projectId)
	resp.Diagnostics.Append(diags...)
	diags = utils.AssertStringNotEmpty("cluster name", clusterId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleting cluster info", map[string]interface{}{
		"cluster": clusterId,
		"project": projectId,
	})

	_, err := utils.GetCluster(ctx, clusterId, projectId, auth)
	if err != nil && err != utils.ErrResourceNotExists {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to get cluster %s in project %s",
			clusterId, projectId), err.Error())
		return
	}

	err = utils.DeleteCluster(ctx, clusterId, projectId, auth)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete cluster %s in project %s",
			clusterId, projectId), err.Error())
		return
	}
}
