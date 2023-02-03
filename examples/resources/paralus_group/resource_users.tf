# This example shows how to create a group resource for user access

resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    users = ["john.smith@example.com"]
}