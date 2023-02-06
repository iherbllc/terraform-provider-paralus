# This example shows how to create a Group resource for project/namespace role access

resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    project_roles {
        group = "test"
        role = "ADMIN"
    }
}