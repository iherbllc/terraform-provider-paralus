package structs

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KubeConfig struct {
	Name                  types.String `tfsdk:"name"`
	Namespace             types.String `tfsdk:"namespace"`
	Cluster               types.String `tfsdk:"cluster"`
	ClusterInfo           types.List   `tfsdk:"cluster_info"`
	ClientCertificateData types.String `tfsdk:"client_certificate_data"`
	ClientKeyData         types.String `tfsdk:"client_key_data"`
}

type ClusterInfo struct {
	CertificateAuthorityData types.String `tfsdk:"certificate_authority_data"`
	Server                   types.String `tfsdk:"server"`
}

func (c ClusterInfo) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"certificate_authority_data": types.StringType,
		"server":                     types.StringType,
	}
}
