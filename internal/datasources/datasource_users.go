// Users Terraform DataSource
package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/iherbllc/terraform-provider-paralus/internal/utils"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = (*DsUsers)(nil)

func DataSourceUsers() datasource.DataSource {
	return &DsUsers{}
}

type DsUsers struct {
	cfg *config.Config
}

func (d *DsUsers) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Paralus DataSource Users
func (d *DsUsers) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information on all paralus users or a filtered few. Uses the [pctl](https://github.com/paralus/cli) library",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Number of users to return. Specify -1 for all (Default: 10)",
				Optional:            true,
			},
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Where to begin the return based on the total number of users (Default: 0)",
				Optional:            true,
			},
			"users_info": schema.ListNestedAttribute{
				MarkdownDescription: "Users information",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"first_name": schema.StringAttribute{
							MarkdownDescription: "User's first name",
							Computed:            true,
							Sensitive:           true,
						},
						"last_name": schema.StringAttribute{
							MarkdownDescription: "User's last name",
							Computed:            true,
							Sensitive:           true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "User's email",
							Computed:            true,
							Sensitive:           true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "User's ID",
							Computed:            true,
						},
						"groups": schema.ListAttribute{
							MarkdownDescription: "List of groups user belong's to",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"project_roles": schema.ListNestedAttribute{
							MarkdownDescription: "Project roles attached to user, containing group or namespace",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"project": schema.StringAttribute{
										Computed: true,
									},
									"role": schema.StringAttribute{
										Computed: true,
									},
									"namespace": schema.StringAttribute{
										Computed: true,
									},
									"group": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filters": schema.SingleNestedBlock{
				MarkdownDescription: "Filters to narrow returned user information",
				Attributes: map[string]schema.Attribute{
					"project": schema.StringAttribute{
						MarkdownDescription: "Name of a project to filter against, as defined under spec.projectNamespaceRoles.project",
						Optional:            true,
					},
					"role": schema.StringAttribute{
						MarkdownDescription: "Name of a role to filter against, as defined under spec.projectNamespaceRoles.role",
						Optional:            true,
					},
					"group": schema.StringAttribute{
						MarkdownDescription: "Name of one of the groups to filter against, as defined within the spec.groups list",
						Optional:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Filter by the user's email address, as defined under metadata.name",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^.*@.*$`),
								"email must be in format: XXXX@XXX.XXX",
							),
						},
					},
					"first_name": schema.StringAttribute{
						MarkdownDescription: "Filter by the user's first name, as defined under spec.firstName",
						Optional:            true,
					},
					"last_name": schema.StringAttribute{
						MarkdownDescription: "Filter by the user's last name, as defined under spec.lastName",
						Optional:            true,
					},
					"case_sensitive": schema.BoolAttribute{
						MarkdownDescription: "Whether to make the filter on first_name, last_name, or email address case-sensitive. (Default: false)",
						Optional:            true,
					},
					"allow_more_than_one": schema.BoolAttribute{
						MarkdownDescription: "Whether to allow more than one record to return when filtering. (Default: false)",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (d *DsUsers) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.cfg = cfg
}

// Retreive Users JSON info
func (d *DsUsers) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.cfg == nil {
		resp.Diagnostics.AddError(
			"Unconfigured PCTL Config",
			"Expected configured PCTL config. Please ensure the values are passed in or report this issue to the provider developers.",
		)
		return
	}

	var data *structs.User
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := data.Limit.ValueInt64()
	if limit == 0 {
		limit = 10
	}
	offset := data.Offset.ValueInt64()

	// build parameters based on the filter given
	params := make([]string, 10)
	params[0] = fmt.Sprintf("organization=%s", d.cfg.Organization)
	params[1] = fmt.Sprintf("partner=%s", d.cfg.Partner)
	params[2] = fmt.Sprintf("limit=%d", limit)
	params[3] = fmt.Sprintf("offset=%d", offset)
	filtersCounter := 4
	filter_is := ""
	filter_is_value := ""
	filter_is_case_senstive := false
	filter_allow_more_than_one := false
	if !data.Filters.IsNull() {
		var filters structs.Filter
		diags := data.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !filters.Role.IsNull() {
			params[filtersCounter] = fmt.Sprintf("role=%s&", filters.Role.ValueString())
			filtersCounter++
		}
		if !filters.Project.IsNull() {
			params[filtersCounter] = fmt.Sprintf("project=%s&", filters.Project.ValueString())
			filtersCounter++
		}
		if !filters.Group.IsNull() {
			params[filtersCounter] = fmt.Sprintf("group=%s&", filters.Group.ValueString())
			filtersCounter++
		}
		filter_is_case_senstive = filters.CaseSensitive.ValueBool()
		filter_allow_more_than_one = filters.AllowMoreThanOne.ValueBool()
		email := filters.Email.ValueString()
		first_name := filters.FirstName.ValueString()
		last_name := filters.LastName.ValueString()
		if email != "" && first_name != "" || email != "" && last_name != "" || first_name != "" && last_name != "" {
			resp.Diagnostics.AddError("Please specify only one: email, first_name, or last_name", "")
			return
		}
		if email != "" {
			params[filtersCounter] = fmt.Sprintf("q=%s&", email)
			filter_is = "email"
			filter_is_value = email
			filtersCounter++
		}
		if first_name != "" {
			params[filtersCounter] = fmt.Sprintf("q=%s&", first_name)
			filter_is = "first_name"
			filter_is_value = first_name
			filtersCounter++
		}
		if last_name != "" {
			params[filtersCounter] = fmt.Sprintf("q=%s&", last_name)
			filter_is = "last_name"
			filter_is_value = last_name
			filtersCounter++
		}
	}

	params = params[0:filtersCounter] // compress the list and remove empty params

	auth := d.cfg.GetAppAuthProfile()
	usersInfo, err := utils.GetUsers(ctx, params, auth)
	if err != nil {
		resp.Diagnostics.AddError("error locating users based on provided values", err.Error())
		return
	}

	if len(usersInfo) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("No users returned based on provided query params: %s", strings.Join(params, "&")), "")
		return
	}

	if val, ok := os.LookupEnv("TF_LOG"); ok && val == "TRACE" {
		for _, userInfo := range usersInfo {
			var usersMap map[string]interface{}
			usersData, err := json.Marshal(userInfo)
			if err != nil {
				resp.Diagnostics.AddError("error converting usersInfo to usersData bytes", err.Error())
				return
			}
			err = json.Unmarshal(usersData, &usersMap)
			if err != nil {
				resp.Diagnostics.AddError("error converting usersData bytes to usersMap map[string]interface{}", err.Error())
				return
			}
			tflog.Trace(ctx, "users returned based on query parameters: ", usersMap)
		}
	}

	// this means that a request to filter on email, first_name or last_name was specified
	if filter_is != "" {
		usersInfo, err = utils.FilterUsers(usersInfo, filter_is, filter_is_value, filter_is_case_senstive, filter_allow_more_than_one)
		if err != nil {
			resp.Diagnostics.AddError("error locating filtered user", err.Error())
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("datasourceUsersRead provider config used: %s", utils.GetConfigAsMap(d.cfg)))

	utils.BuildResourceFromUsersStruct(ctx, usersInfo, data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
