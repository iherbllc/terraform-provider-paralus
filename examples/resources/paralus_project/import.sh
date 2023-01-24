# Import an existing project into TF
# Format should be: terraform import paralus_project.<RESOURCE_NAME> <PROJECT_NAME>
#
# NOTE: Project must exist and provider must have designated Partner and Organization value or request will fail

terraform import paralus_project.test myproject