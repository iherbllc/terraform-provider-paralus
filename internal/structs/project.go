package structs

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Project struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Uuid         types.String `tfsdk:"uuid"`
	ProjectRoles types.List   `tfsdk:"project_roles"`
	UserRoles    types.List   `tfsdk:"user_roles"`
}

type UserRole struct {
	User      types.String `tfsdk:"user"`
	Role      types.String `tfsdk:"role"`
	Namespace types.String `tfsdk:"namespace"`
}

func (u UserRole) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"user":      types.StringType,
		"role":      types.StringType,
		"namespace": types.StringType,
	}
}
