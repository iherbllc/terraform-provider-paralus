package structs

import "github.com/hashicorp/terraform-plugin-framework/types"

type Group struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	ProjectRoles types.List   `tfsdk:"project_roles"`
	Users        types.List   `tfsdk:"users"`
	Type         types.String `tfsdk:"type"`
}
