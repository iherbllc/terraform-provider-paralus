# This example shows how to create a Group resource for project/namespace role access
# with multiple project_role blocks
resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    project_roles {
        group = "group1"
        role = "PROJECT_ADMIN"
        project = "project1"
    }
    project_roles {
        group = "group2"
        role = "NAMESPACE_ADMIN"
        namespace = "platform"
        project = "project2"
    }
    project_roles {
        group = "group2"
        role = "NAMESPACE_READ_ONLY"
        namesapce = "monitoring"
        project = "test2"
    }
    project_roles {
        group = "group3"
        role = "NAMESPACE_ADMIN"
        namesapce = "prometheus"
        project = "test3"
    }
}