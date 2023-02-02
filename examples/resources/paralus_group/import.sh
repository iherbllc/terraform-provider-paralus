# Import an existing group into TF
# Format should be: terraform import paralu_group.<RESOURCE_NAME> <GROUP_NAME>
#
# NOTE: Group must exist and provider must have designated Partner and Organization value or request will fail

terraform import paralus_group.test myproject