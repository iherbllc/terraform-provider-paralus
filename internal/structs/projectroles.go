package structs

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectRole struct {
	Project   types.String `tfsdk:"project"`
	Role      types.String `tfsdk:"role"`
	Namespace types.String `tfsdk:"namespace"`
	Group     types.String `tfsdk:"group"`
}

func (p ProjectRole) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"project":   types.StringType,
		"role":      types.StringType,
		"namespace": types.StringType,
		"group":     types.StringType,
	}
}
