# This example shows how to create a project resource

resource "paralus_project" "test" {
    name = "test"
    description = "test project"
    user_roles {
        user_id = "john.smith@example.com"
        role = "PROJECT_ADMIN"
    }
    user_roles {
        user_id = "jane.doe@example.com"
        role = "NAMESPACE_ADMIN"
        namespace = "platform"
    }
}