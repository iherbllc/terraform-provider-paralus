# Import an existing cluster into TF
# Format should be: terraform import paralus_cluster.<RESOURCE_NAME> <PROJECT_NAME>:<CLUSTER_NAME>
#
# NOTE: Both project and cluster must exist or the request will fail

terraform import paralus_cluster.test myproject:mycluster