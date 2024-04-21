package structs

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Cluster struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ClusterType    types.String `tfsdk:"cluster_type"`
	Uuid           types.String `tfsdk:"uuid"`
	Params         types.Object `tfsdk:"params"`
	Project        types.String `tfsdk:"project"`
	BSFileCombined types.String `tfsdk:"bootstrap_files_combined"`
	BSFiles        types.List   `tfsdk:"bootstrap_files"`
	Labels         types.Map    `tfsdk:"labels"`
	Annotations    types.Map    `tfsdk:"annotations"`
	Relays         types.String `tfsdk:"relays"`
}

type Params struct {
	ProvisionType        types.String `tfsdk:"provision_type"`
	ProvisionEnvironment types.String `tfsdk:"provision_environment"`
	ProvisionPackageType types.String `tfsdk:"provision_package_type"`
	EnvironmentProvider  types.String `tfsdk:"environment_provider"`
	KubernetesProvider   types.String `tfsdk:"kubernetes_provider"`
	State                types.String `tfsdk:"state"`
}

func (p Params) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provision_type":         types.StringType,
		"provision_environment":  types.StringType,
		"provision_package_type": types.StringType,
		"environment_provider":   types.StringType,
		"kubernetes_provider":    types.StringType,
		"state":                  types.StringType,
	}
}
