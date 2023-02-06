# This example shows how to create a project resource

resource "paralus_project" "test" {
    name = "test"
    description = "test project"
    project_roles {
        group = "group1"
        role = "PROJECT_ADMIN"
        project = "test"
    }
    project_roles {
        group = "group2"
        role = "NAMESPACE_ADMIN"
        namespace = "platform"
        project = "test"
    }
    project_roles {
        group = "group2"
        role = "NAMESPACE_READ_ONLY"
        namespace = "monitoring"
        project = "test"
    }
    project_roles {
        group = "group3"
        role = "NAMESPACE_ADMIN"
        namespace = "prometheus"
        project = "test"
    }
}