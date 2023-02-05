---
page_title: "Resource - paralus_group"
subcategory: ""
description: |-
  Resource containing paralus group information. Uses the [pctl](https://github.com/paralus/cli) library
---

# paralus_group (Resource)

Resource containing paralus group information. Uses the [pctl](https://github.com/paralus/cli) library


## Example Usage

### Roles - Single

```terraform
# This example shows how to create a Group resource for project/namespace role access

resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    project_roles {
        group = "test"
        role = "ADMIN"
    }
}
```

### Roles - Multiple

```terraform
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
```

### Users

```terraform
# This example shows how to create a group resource for user access

resource "paralus_group" "test" {
    name = "test"
    description = "test group"
    users = ["john.smith@example.com"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Group name

### Optional

- `description` (String) Group description.
- `project_roles` (Block List) Project namespace roles to attach to the group (see [below for nested schema](#nestedblock--project_roles))
- `users` (List of String) User roles attached to group

### Read-Only

- `id` (String) Group ID in the format "GROUP_NAME"
- `type` (String) Type of group

<a id="nestedblock--project_roles"></a>
### Nested Schema for `project_roles`

Required:

- `group` (String) Authorized group
- `role` (String) Role name

Optional:

- `namespace` (String) Authorized namespace
- `project` (String) Project name