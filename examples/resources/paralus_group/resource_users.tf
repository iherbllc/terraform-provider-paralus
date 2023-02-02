# This example shows how to create a project resource

resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    users = ["john.smith@example.com"]
}