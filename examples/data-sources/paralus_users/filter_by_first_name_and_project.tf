data "paralus_users" "default" {
  filter {
    first_name = "Test"
    project = "default"
  }
}