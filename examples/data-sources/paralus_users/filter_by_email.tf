data "paralus_users" "default" {
  filter {
    email = "test@test.com"
  }
}