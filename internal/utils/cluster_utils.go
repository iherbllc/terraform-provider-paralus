// Utility methods for PCTL Cluster manipulation
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iherbllc/terraform-provider-paralus/internal/structs"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8Scheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/paralus/cli/pkg/authprofile"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

// Build the cluster struct from a schema resource
func BuildClusterStructFromResource(ctx context.Context, data *structs.Cluster) (*infrav3.Cluster, diag.Diagnostics) {

	clusterStruct := &infrav3.Cluster{
		Kind: "Cluster",
		Metadata: &commonv3.Metadata{
			Name:    data.Name.ValueString(),
			Project: data.Project.ValueString(),
		},
		Spec: &infrav3.ClusterSpec{
			Metro:       &infrav3.Metro{},
			ClusterType: data.ClusterType.ValueString(),
		},
	}

	// If we have params, let's add them into the struct
	if !data.Params.IsNull() {
		var params structs.Params
		diags := data.Params.As(ctx, &params, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		provisionParams := &infrav3.ProvisionParams{
			EnvironmentProvider:  params.EnvironmentProvider.ValueString(),
			KubernetesProvider:   params.KubernetesProvider.ValueString(),
			ProvisionEnvironment: params.ProvisionEnvironment.ValueString(),
			ProvisionPackageType: params.ProvisionPackageType.ValueString(),
			ProvisionType:        params.ProvisionType.ValueString(),
			State:                params.State.ValueString(),
		}

		clusterStruct.Spec.Params = provisionParams
	}

	if !data.Labels.IsNull() {
		if clusterStruct.Metadata.Labels == nil {
			clusterStruct.Metadata.Labels = make(map[string]string)
		}
		labels := make(map[string]types.String, len(data.Labels.Elements()))
		diags := data.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			return nil, diags
		}
		for k, v := range labels {
			clusterStruct.Metadata.Labels[k] = v.ValueString()
		}
	}

	if !data.Annotations.IsNull() {
		if clusterStruct.Metadata.Annotations == nil {
			clusterStruct.Metadata.Annotations = make(map[string]string)
		}
		annotations := make(map[string]types.String, len(data.Annotations.Elements()))
		diags := data.Annotations.ElementsAs(ctx, &annotations, false)
		if diags.HasError() {
			return nil, diags
		}
		for k, v := range annotations {
			clusterStruct.Metadata.Annotations[k] = v.ValueString()
		}
	}

	return clusterStruct, nil
}

// Build the schema resource from Cluster Struct
func BuildResourceFromClusterStruct(ctx context.Context, cluster *infrav3.Cluster, data *structs.Cluster, auth *authprofile.Profile) diag.Diagnostics {
	var diagsReturn diag.Diagnostics
	var diags diag.Diagnostics
	data.Id = types.StringValue(cluster.Metadata.Project + ":" + cluster.Metadata.Name) // will be removed eventually
	data.Name = types.StringValue(cluster.Metadata.Name)
	data.Description = types.StringValue(cluster.Metadata.Description)
	data.Project = types.StringValue(cluster.Metadata.Project)
	data.ClusterType = types.StringValue(cluster.Spec.ClusterType)
	data.Uuid = types.StringValue(cluster.Metadata.Id)
	if cluster.Spec.Params != nil {
		envProvider := types.StringValue(cluster.Spec.Params.EnvironmentProvider)
		if envProvider == types.StringValue("") {
			envProvider = types.StringNull()
		}
		provisionEnv := types.StringValue(cluster.Spec.Params.ProvisionEnvironment)
		if provisionEnv == types.StringValue("") {
			provisionEnv = types.StringNull()
		}
		k8sProvider := types.StringValue(cluster.Spec.Params.KubernetesProvider)
		if k8sProvider == types.StringValue("") {
			k8sProvider = types.StringNull()
		}
		provisionPkgType := types.StringValue(cluster.Spec.Params.ProvisionPackageType)
		if provisionPkgType == types.StringValue("") {
			provisionPkgType = types.StringNull()
		}
		provisionType := types.StringValue(cluster.Spec.Params.ProvisionType)
		if provisionType == types.StringValue("") {
			provisionType = types.StringNull()
		}
		state := types.StringValue(cluster.Spec.Params.State)
		if state == types.StringValue("") {
			state = types.StringNull()
		}
		params := structs.Params{
			EnvironmentProvider:  envProvider,
			KubernetesProvider:   k8sProvider,
			ProvisionEnvironment: provisionEnv,
			ProvisionPackageType: provisionPkgType,
			ProvisionType:        provisionType,
			State:                state,
		}
		data.Params, diags = types.ObjectValueFrom(ctx, params.AttributeTypes(), params)
		diagsReturn.Append(diags...)

	}

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, cluster.Metadata.Labels)
	diagsReturn.Append(diags...)
	data.Annotations, diags = types.MapValueFrom(ctx, types.StringType, cluster.Metadata.Annotations)
	diagsReturn.Append(diags...)

	relays, bsfiles, bsfile, err := SetBootstrapFileAndRelays(ctx, cluster.Metadata.Project, cluster.Metadata.Name, auth)
	if err != nil {
		diagsReturn.AddError("Setting bootstrap file and relays failed", err.Error())
	} else {
		data.Relays = types.StringValue(relays)
		data.BSFiles, diags = types.ListValueFrom(ctx, types.StringType, bsfiles)
		diagsReturn.Append(diags...)
		data.BSFileCombined = types.StringValue(bsfile)
	}

	return diagsReturn
}

// Splits a single YAML file containing multiple YAML entries into a list of string
func splitSingleYAMLIntoList(singleYAML string) []string {
	docs := strings.Split(string(singleYAML), "\n---")

	yamls := []string{}
	// Trim whitespace in both ends of each yaml docs.
	// - Re-add a single newline last
	for _, doc := range docs {
		content := strings.TrimSpace(doc)
		// Ignore empty docs
		if content != "" {
			yamls = append(yamls, content)
		}
	}
	return yamls
}

// Retrieve the relays info from the bootstrap files
// Which are found within the relay-agent-config configmap
func getBootstrapRelays(bootstrapFiles []string) (string, error) {
	// yamlFiles is an []string
	for _, boostrapFile := range bootstrapFiles {

		decode := k8Scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(boostrapFile), nil, nil)

		if err != nil {
			return boostrapFile, err
		}

		// now use switch over the type of the object
		// and match each type-case
		switch o := obj.(type) {
		// case *v1.Pod:
		// 	// o is a pod
		// case *v1beta1.Role:
		// 	// o is the actual role Object with all fields etc
		// case *v1beta1.RoleBinding:
		// case *v1beta1.ClusterRole:
		// case *v1beta1.ClusterRoleBinding:
		// case *v1.ServiceAccount:
		case *v1.ConfigMap:
			targetConfigMap := o.Data
			if relays, ok := targetConfigMap["relays"]; ok {
				return relays, nil
			}
		default:
			//o is unknown for us
		}
	}
	return "", nil
}

// Retrieve the YAML files that will be used to setup paralus agents in cluster and assign it to the schema
// Also retrieve the relays from  the data of the relay-agent configMap YAML
// Due to the parallel nature of testing, it might be that the cluster would be created
// before the relay was effectively populated. So let's do a increased delay check
func SetBootstrapFileAndRelays(ctx context.Context, projectId, clusterId string,
	auth *authprofile.Profile) (string, []string, string, error) {

	b := &backoff.Backoff{
		Jitter: true,
		Max:    5 * time.Minute,
	}

	var bootstrapFile string
	var bootstrapFiles []string
	var relay_or_resp string
	var err error

	for {
		// already checked earlier for cluster to exist, so don't have to check again.
		bootstrapFile, err = GetBootstrapFile(ctx, clusterId, projectId, auth)

		if err != nil {
			return "", nil, "", errors.Wrapf(err, "Error retrieving bootstrap file for cluster %s in project %s",
				clusterId, projectId)
		}

		bootstrapFiles = splitSingleYAMLIntoList(bootstrapFile)
		relay_or_resp, err = getBootstrapRelays(bootstrapFiles)
		if err != nil {
			return "", nil, "", errors.Wrapf(err, "Error while decoding YAML object %s", relay_or_resp)
		}

		d := b.Duration()
		if relay_or_resp != "" || d >= b.Max {
			break
		}

		// If the GetBootstrapFile call is too fast, it might lead to the relay info not ending up in the cluster
		// yet, so will need to try again. Using jitter to avoid flooding the API.

		tflog.Info(ctx, fmt.Sprintf("No relay populated yet, retrying in %s", d))
		time.Sleep(d)
	}

	if relay_or_resp == "" {
		return "", nil, "", errors.Errorf("Unable to retrieve relay info from created cluster within %s", b.Duration())
	}

	b.Reset()

	return relay_or_resp, bootstrapFiles, bootstrapFile, nil
}

// Will retrieve the bootstrap file for imported clusters
func GetBootstrapFile(ctx context.Context, name, project string, auth *authprofile.Profile) (string, error) {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s/download", project, name)
	return makeRestCall(ctx, uri, "GET", nil, auth)

}

// Retrieves cluster info
func GetCluster(ctx context.Context, name, project string, auth *authprofile.Profile) (*infrav3.Cluster, error) {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", project, name)
	resp, err := makeRestCall(ctx, uri, "GET", nil, auth)
	if err != nil {
		return nil, err
	}
	var cluster infrav3.Cluster
	if err := json.Unmarshal([]byte(resp), &cluster); err != nil {
		return nil, fmt.Errorf("error unmarshalling cluster details: %s", err)
	}
	return &cluster, nil
}

// Delete the cluster
func DeleteCluster(ctx context.Context, name, project string, auth *authprofile.Profile) error {
	// get cluster
	_, err := GetCluster(ctx, name, project, auth)

	if err == ErrResourceNotExists {
		return nil
	}

	if err != nil {
		return err
	}

	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", project, name)
	_, err = makeRestCall(ctx, uri, "DELETE", nil, auth)
	if err != nil {
		return err
	}

	return nil
}

// Update cluster takes the updated cluster details and sends it to the core
func CreateCluster(ctx context.Context, cluster *infrav3.Cluster, auth *authprofile.Profile) error {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster", cluster.Metadata.Project)
	resp, err := makeRestCall(ctx, uri, "POST", cluster, auth)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp), &cluster); err != nil {
		return fmt.Errorf("error unmarshalling cluster details: %s", err)
	}
	return nil
}

// Update cluster takes the updated cluster details and sends it to the core
func UpdateCluster(ctx context.Context, cluster *infrav3.Cluster, auth *authprofile.Profile) error {
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster/%s", cluster.Metadata.Project, cluster.Metadata.Name)
	resp, err := makeRestCall(ctx, uri, "PUT", cluster, auth)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp), &cluster); err != nil {
		return fmt.Errorf("error unmarshalling cluster details: %s", err)
	}
	return nil
}

// ListAllClusters uses the lower level func ListClusters to retrieve a list of all clusters
func ListAllClusters(ctx context.Context, projectId string, auth *authprofile.Profile) ([]*infrav3.Cluster, error) {
	var clusters []*infrav3.Cluster
	limit := 10000
	c, count, err := listClusters(ctx, projectId, limit, 0, auth)
	if err != nil {
		return nil, err
	}
	clusters = c
	for count > limit {
		offset := limit
		limit = count
		c, _, err = listClusters(ctx, projectId, limit, offset, auth)
		if err != nil {
			return clusters, err
		}
		clusters = append(clusters, c...)
	}
	return clusters, nil
}

/*
ListClusters paginates through a list of clusters
*/
func listClusters(ctx context.Context, project string, limit, offset int, auth *authprofile.Profile) ([]*infrav3.Cluster, int, error) {
	// check to make sure the limit or offset is not negative
	if limit < 0 || offset < 0 {
		return nil, 0, fmt.Errorf("provided limit (%d) or offset (%d) cannot be negative", limit, offset)
	}
	uri := fmt.Sprintf("/infra/v3/project/%s/cluster?limit=%d&offset=%d", project, limit, offset)
	resp, err := makeRestCall(ctx, uri, "GET", nil, auth)
	if err != nil {
		return nil, 0, err
	}
	a := infrav3.ClusterList{}
	_ = json.Unmarshal([]byte(resp), &a)
	return a.Items, int(a.Metadata.Count), nil
}
