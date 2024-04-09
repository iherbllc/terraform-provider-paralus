---
page_title: "Data Source - paralus_users"
subcategory: ""
description: |-
  Retrieves all/filtered users information. Uses the [pctl](https://github.com/paralus/cli) library
---

# paralus_users (Data Source)

Retrieves all/filtered users information. Uses the [pctl](https://github.com/paralus/cli) library

## Example Usage

### The first 10 Users' Information

```terraform
data "paralus_users" "default" {
}
```

### All Users' Information

```terraform
data "paralus_users" "default" {
  limit = -1
}
```

### Users 10-20 Information

```terraform
data "paralus_users" "default" {
  offset = 10
  limit = 10
}
```

### Filtered User's Information (by email)

```terraform
data "paralus_users" "default" {
  filter {
    email = "test@test.com"
  }
}
```

Note: A similar request can be made with `first_name` and `last_name`. However, only one of the three can be used to filter. Project, Role, and Group can be included. For example,

```terraform
data "paralus_users" "default" {
  filter {
    first_name = "Test"
    project = "default"
  }
}
```
