// Users Terraform DataSource
package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iherbllc/terraform-provider-paralus/internal/utils"
	"github.com/pkg/errors"

	"github.com/paralus/cli/pkg/config"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Paralus DataSource Users
func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves all/filtered users information. Uses the [pctl](https://github.com/paralus/cli) library",
		ReadContext: datasourceUsersRead,
		Schema: map[string]*schema.Schema{
			"limit": {
				Type:        schema.TypeInt,
				Description: "Number of users to return. Specify -1 for all",
				Optional:    true,
				Default:     10,
			},
			"offset": {
				Type:        schema.TypeInt,
				Description: "Where to begin the return based on the total number of users",
				Optional:    true,
				Default:     0,
			},
			"users_info": {
				Type:        schema.TypeList,
				Description: "Users information",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"first_name": {
							Type:        schema.TypeString,
							Description: "User's first name",
							Computed:    true,
							Sensitive:   true,
						},
						"last_name": {
							Type:        schema.TypeString,
							Description: "User's last name",
							Computed:    true,
							Sensitive:   true,
						},
						"email": {
							Type:        schema.TypeString,
							Description: "User's email",
							Computed:    true,
							Sensitive:   true,
						},
						"id": {
							Type:        schema.TypeString,
							Description: "User's ID",
							Computed:    true,
						},
						"groups": {
							Type:        schema.TypeList,
							Description: "List of groups user belong's to",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"project_roles": {
							Type:        schema.TypeList,
							Description: "Project roles attached to user, containing group or namespace",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"project": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"role": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"group": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"filters": {
				Type:        schema.TypeList,
				Description: "Filters to narrow returned user information",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"email": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: func(val any, key string) (warns []string, errs []error) {
								v := val.(string)
								if !regexp.MustCompile(`^.*@.*$`).MatchString(v) {
									errs = append(errs, fmt.Errorf("%q must be in format: XXXX@XXX.XXX, got: %s", key, v))
								}
								return
							},
						},
						"first_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"last_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"case_sensitive": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"allow_more_than_one": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

// Retreive Users JSON info
func datasourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	cfg := m.(*config.Config)
	auth := cfg.GetAppAuthProfile()

	limit := d.Get("limit").(int)
	offset := d.Get("offset").(int)

	var id strings.Builder
	fmt.Fprintf(&id, "%d%d", limit, offset)

	// build parameters based on the filter given
	params := make([]string, 10)
	params[0] = fmt.Sprintf("organization=%s", cfg.Organization)
	params[1] = fmt.Sprintf("partner=%s", cfg.Partner)
	params[2] = fmt.Sprintf("limit=%d", limit)
	params[3] = fmt.Sprintf("offset=%d", offset)
	filtersCounter := 4
	filter_is := ""
	filter_is_value := ""
	filter_is_case_senstive := false
	filter_allow_more_than_one := false
	if filtersAttr, ok := d.GetOk("filters"); ok {
		filters := filtersAttr.([]interface{})
		for _, eachFilter := range filters {
			if filter, ok := eachFilter.(map[string]interface{}); ok {
				role := filter["role"]
				if role != "" {
					params[filtersCounter] = fmt.Sprintf("role=%s&", role)
					id.WriteString(role.(string))
					filtersCounter++
				}
				project := filter["project"]
				if project != "" {
					params[filtersCounter] = fmt.Sprintf("project=%s&", project)
					id.WriteString(project.(string))
					filtersCounter++
				}
				group := filter["group"]
				if group != "" {
					params[filtersCounter] = fmt.Sprintf("group=%s&", group)
					id.WriteString(group.(string))
					filtersCounter++
				}
				filter_is_case_senstive = filter["case_sensitive"].(bool)
				filter_allow_more_than_one = filter["allow_more_than_one"].(bool)
				email := filter["email"]
				first_name := filter["first_name"]
				last_name := filter["last_name"]
				if email != "" && first_name != "" || email != "" && last_name != "" || first_name != "" && last_name != "" {
					return diag.Errorf("Please specify only one: email, first_name, or last_name")
				}
				if email != "" {
					params[filtersCounter] = fmt.Sprintf("q=%s&", email)
					id.WriteString(email.(string))
					filter_is = "email"
					filter_is_value = email.(string)
					filtersCounter++
				}
				if first_name != "" {
					params[filtersCounter] = fmt.Sprintf("q=%s&", first_name)
					id.WriteString(first_name.(string))
					filter_is = "first_name"
					filter_is_value = first_name.(string)
					filtersCounter++
				}
				if last_name != "" {
					params[filtersCounter] = fmt.Sprintf("q=%s&", last_name)
					id.WriteString(last_name.(string))
					filter_is = "last_name"
					filter_is_value = last_name.(string)
					filtersCounter++
				}
			}
		}
	}

	params = params[0:filtersCounter] // compress the list and remove empty params

	usersInfo, err := utils.GetUsers(ctx, params, auth)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "error locating users based on provided values"))
	}

	if len(usersInfo) == 0 {
		return diag.Errorf("No users returned based on provided query params: %s", strings.Join(params, "&"))
	}

	if val, ok := os.LookupEnv("TF_LOG"); ok && val == "TRACE" {
		for _, userInfo := range usersInfo {
			var usersMap map[string]interface{}
			usersData, err := json.Marshal(userInfo)
			if err != nil {
				return diag.FromErr(errors.Wrapf(err, "error converting usersInfo to usersData bytes"))
			}
			err = json.Unmarshal(usersData, &usersMap)
			if err != nil {
				return diag.FromErr(errors.Wrapf(err, "error converting usersData bytes to usersMap map[string]interface{}"))
			}
			tflog.Trace(ctx, "users returned based on query parameters: ", usersMap)
		}
	}

	// this means that a request to filter on email, first_name or last_name was specified
	if filter_is != "" {
		usersInfo, err = utils.FilterUsers(usersInfo, filter_is, filter_is_value, filter_is_case_senstive, filter_allow_more_than_one)
		if err != nil {
			return diag.FromErr(errors.Wrapf(err, "error locating filtered user"))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("datasourceUsersRead provider config used: %s", utils.GetConfigAsMap(cfg)))

	utils.BuildResourceFromUsersStruct(usersInfo, d)

	d.SetId(id.String()) // ID is a value made up of the offset,limit,and whatever specified filters

	return diags

}
