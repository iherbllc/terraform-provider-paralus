# This example shows how to create a project resource

resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    project_roles {
        group = "test"
        role = "ADMIN"
    }
}