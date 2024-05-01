// Utility methods for PCTL Group struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iherbllc/terraform-provider-paralus/internal/structs"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	groupv3 "github.com/paralus/paralus/proto/types/userpb/v3"
)

// Build the group struct from a schema resource
func BuildGroupStructFromResource(ctx context.Context, data *structs.Group) (*groupv3.Group, diag.Diagnostics) {

	groupStruct := groupv3.Group{
		Kind: "Group",
		Metadata: &commonv3.Metadata{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
		},
		Spec: &groupv3.GroupSpec{},
	}

	if !data.ProjectRoles.IsNull() {
		projectRoles := make([]structs.ProjectRole, 0, len(data.ProjectRoles.Elements()))
		diags := data.ProjectRoles.ElementsAs(ctx, &projectRoles, false)
		if diags.HasError() {
			return nil, diags
		}
		groupStruct.Spec.ProjectNamespaceRoles = make([]*groupv3.ProjectNamespaceRole, 0)
		group := data.Name.ValueString() // group will always default to the group name to avoid user error
		for _, projectRole := range projectRoles {
			namespace := projectRole.Namespace.ValueString()
			project := projectRole.Project.ValueString()
			groupStruct.Spec.ProjectNamespaceRoles = append(groupStruct.Spec.ProjectNamespaceRoles, &groupv3.ProjectNamespaceRole{
				Project:   &project,
				Role:      projectRole.Role.ValueString(),
				Namespace: &namespace,
				Group:     &group,
			})
		}
	}

	if !data.Users.IsNull() {
		users := make([]types.String, 0, len(data.Users.Elements()))
		diags := data.Users.ElementsAs(ctx, &users, false)
		if diags.HasError() {
			return nil, diags
		}
		groupStruct.Spec.Users = make([]string, len(users))
		for i, v := range users {
			groupStruct.Spec.Users[i] = v.ValueString()
		}
	}

	groupStruct.Spec.Type = data.Type.ValueString()
	if groupStruct.Spec.Type == "" {
		groupStruct.Spec.Type = "SYSTEM"
	}

	return &groupStruct, nil
}

// Build the schema resource from group Struct
func BuildResourceFromGroupStruct(ctx context.Context, group *groupv3.Group, data *structs.Group) diag.Diagnostics {
	var diagsReturn diag.Diagnostics
	var diags diag.Diagnostics
	data.Id = types.StringValue(group.Metadata.Name)
	data.Name = types.StringValue(group.Metadata.Name)
	data.Description = types.StringValue(group.Metadata.Description)
	projectRoles := make([]structs.ProjectRole, 0)
	for _, role := range group.Spec.GetProjectNamespaceRoles() {
		project := types.StringValue(DerefString(role.Project))
		if project == types.StringValue("") {
			project = types.StringNull()
		}
		namespace := types.StringValue(DerefString(role.Namespace))
		if namespace == types.StringValue("") {
			namespace = types.StringNull()
		}
		roleVal := types.StringValue(role.Role)
		if roleVal == types.StringValue("") {
			roleVal = types.StringNull()
		}
		group := types.StringValue(DerefString(role.Group))
		if group == types.StringValue("") {
			group = types.StringNull()
		}
		projectRoles = append(projectRoles, structs.ProjectRole{
			Project:   project,
			Role:      roleVal,
			Namespace: namespace,
			Group:     group,
		})
	}

	// Couresty of https://github.com/hashicorp/terraform-plugin-framework/issues/713#issuecomment-1577729734
	data.ProjectRoles, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: structs.ProjectRole{}.AttributeTypes()}, projectRoles)
	diagsReturn.Append(diags...)
	data.Users, diags = types.ListValueFrom(ctx, types.StringType, group.Spec.Users)
	diagsReturn.Append(diags...)
	data.Type = types.StringValue(group.Spec.Type)

	return diags
}

// Check groups specified in the ProjectNamespaceRoles struct exist in Paralus
func CheckGroupsFromPNRStructExist(ctx context.Context, pnrStruct []*groupv3.ProjectNamespaceRole, auth *authprofile.Profile) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(pnrStruct) > 0 {
		for _, pnr := range pnrStruct {
			groupName := pnr.Group
			if groupName != nil {
				// error if we have an empty group name
				if *groupName == "" {
					diags.AddError("group name cannot be empty", "")
					return diags
				}
				_, err := GetGroupByName(ctx, *groupName, auth)
				if err != nil {
					if err == ErrResourceNotExists {
						diags.AddError(fmt.Sprintf("group '%s' does not exist", *groupName), "")
						return diags
					}
					diags.AddError(fmt.Sprintf("error getting group %s info", *groupName), err.Error())
					return diags
				}

			}
		}
	}

	return diags
}

// Get group by name
func GetGroupByName(ctx context.Context, groupName string, auth *authprofile.Profile) (*groupv3.Group, error) {
	cfg := config.GetConfig()
	uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/group/%s", cfg.Partner, cfg.Organization, groupName)
	resp, err := makeRestCall(ctx, uri, "GET", nil, auth)
	if err != nil {
		return nil, err
	}
	grp := &groupv3.Group{}
	err = json.Unmarshal([]byte(resp), grp)
	if err != nil {
		return nil, err
	}

	return grp, nil
}

// Apply group takes the group details and sends it to the core
func ApplyGroup(ctx context.Context, grp *groupv3.Group, auth *authprofile.Profile) error {
	cfg := config.GetConfig()
	grpExisting, err := GetGroupByName(ctx, grp.Metadata.Name, auth)
	if grpExisting != nil {
		tflog.Debug(ctx, fmt.Sprintf("updating group: %s", grp.Metadata.Name))
		uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/group/%s", cfg.Partner, cfg.Organization, grp.Metadata.Name)
		resp, err := makeRestCall(ctx, uri, "PUT", grp, auth)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(resp), grp)
		if err != nil {
			return err
		}
	} else {

		if err != nil && err != ErrResourceNotExists {
			return err
		}

		tflog.Debug(ctx, fmt.Sprintf("creating group: %s", grp.Metadata.Name))
		uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/groups", cfg.Partner, cfg.Organization)
		resp, err := makeRestCall(ctx, uri, "POST", grp, auth)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(resp), grp)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete group
func DeleteGroup(ctx context.Context, groupName string, auth *authprofile.Profile) error {
	_, err := GetGroupByName(ctx, groupName, auth)
	if err == ErrResourceNotExists {
		return nil
	}

	if err != nil {
		return err
	}

	cfg := config.GetConfig()
	uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/group/%s", cfg.Partner, cfg.Organization, groupName)
	_, err = makeRestCall(ctx, uri, "DELETE", nil, auth)
	return err
}
