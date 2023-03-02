// Utility methods for PCTL Group struct
package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/paralus/cli/pkg/authprofile"
	"github.com/paralus/cli/pkg/config"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	groupv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/pkg/errors"
)

// Build the group struct from a schema resource
func BuildGroupStructFromResource(d *schema.ResourceData) *groupv3.Group {

	groupStruct := groupv3.Group{
		Kind: "Group",
		Metadata: &commonv3.Metadata{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		},
		Spec: &groupv3.GroupSpec{},
	}

	if projectRoles, ok := d.GetOk("project_roles"); ok {
		groupStruct.Spec.ProjectNamespaceRoles = make([]*groupv3.ProjectNamespaceRole, 0)
		rolesList := projectRoles.([]interface{})
		group := d.Get("name").(string) // group will always default to the group name to avoid user error
		for _, eachRole := range rolesList {
			if role, ok := eachRole.(map[string]interface{}); ok {
				namespace := role["namespace"].(string)
				project := role["project"].(string)
				groupStruct.Spec.ProjectNamespaceRoles = append(groupStruct.Spec.ProjectNamespaceRoles, &groupv3.ProjectNamespaceRole{
					Project:   &project,
					Role:      role["role"].(string),
					Namespace: &namespace,
					Group:     &group,
				})
			}
		}
	}

	if users, ok := d.GetOk("users"); ok {
		usersList := users.([]interface{})
		groupStruct.Spec.Users = make([]string, len(usersList))
		for i, v := range usersList {
			groupStruct.Spec.Users[i] = v.(string)
		}
	}

	if groupType, ok := d.GetOk("type"); ok {
		groupStruct.Spec.Type = groupType.(string)
	}

	return &groupStruct
}

// Build the schema resource from group Struct
func BuildResourceFromGroupStruct(group *groupv3.Group, d *schema.ResourceData) {
	d.Set("name", group.Metadata.Name)
	d.Set("description", group.Metadata.Description)
	projectRoles := make([]map[string]interface{}, 0)
	for _, role := range group.Spec.GetProjectNamespaceRoles() {
		projectRoles = append(projectRoles, map[string]interface{}{
			"project":   role.Project,
			"role":      role.Role,
			"namespace": role.Namespace,
			"group":     role.Group,
		})
	}
	d.Set("project_roles", projectRoles)
	d.Set("users", group.Spec.Users)
	d.Set("type", group.Spec.Type)
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
					return diag.FromErr(errors.New("group name cannot be empty"))
				}
				_, err := GetGroupByName(ctx, *groupName, auth)
				if err != nil {
					if err == ErrResourceNotExists {
						return diag.FromErr(fmt.Errorf("group '%s' does not exist", *groupName))
					}
					return diag.FromErr(errors.Wrapf(err, "error getting group %s info",
						*groupName))
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
		_, err := makeRestCall(ctx, uri, "PUT", grp, auth)
		if err != nil {
			return err
		}
	} else {

		if err != nil && err != ErrResourceNotExists {
			return err
		}

		tflog.Debug(ctx, fmt.Sprintf("creating group: %s", grp.Metadata.Name))
		uri := fmt.Sprintf("/auth/v3/partner/%s/organization/%s/groups", cfg.Partner, cfg.Organization)
		_, err := makeRestCall(ctx, uri, "POST", grp, auth)
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
