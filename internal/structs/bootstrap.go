package structs

import "github.com/hashicorp/terraform-plugin-framework/types"

type BootstrapFileData struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Uuid           types.String `tfsdk:"uuid"`
	Project        types.String `tfsdk:"project"`
	BSFileCombined types.String `tfsdk:"bootstrap_files_combined"`
	BSFiles        types.List   `tfsdk:"bootstrap_files"`
	Relays         types.String `tfsdk:"relays"`
}
