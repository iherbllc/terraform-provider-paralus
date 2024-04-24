package structs

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type User struct {
	Limit     types.Int64  `tfsdk:"limit"`
	Offset    types.Int64  `tfsdk:"offset"`
	UsersInfo types.List   `tfsdk:"users_info"`
	Filters   types.Object `tfsdk:"filters"`
}

type UserInfo struct {
	FirstName    types.String `tfsdk:"first_name"`
	LastName     types.String `tfsdk:"last_name"`
	Email        types.String `tfsdk:"email"`
	Id           types.String `tfsdk:"id"`
	Groups       types.List   `tfsdk:"groups"`
	ProjectRoles types.List   `tfsdk:"project_roles"`
}

func (u UserInfo) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"first_name":    types.StringType,
		"last_name":     types.StringType,
		"email":         types.StringType,
		"id":            types.StringType,
		"groups":        types.ListType{ElemType: types.StringType},
		"project_roles": types.ListType{ElemType: types.ObjectType{AttrTypes: ProjectRole{}.AttributeTypes()}},
	}
}

type Filter struct {
	Project          types.String `tfsdk:"project"`
	Role             types.String `tfsdk:"role"`
	Group            types.String `tfsdk:"group"`
	Email            types.String `tfsdk:"email"`
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
	CaseSensitive    types.Bool   `tfsdk:"case_sensitive"`
	AllowMoreThanOne types.Bool   `tfsdk:"allow_more_than_one"`
}
