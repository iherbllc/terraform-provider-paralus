# This example shows a bootstrap file data source

data "paralus_bootstrap_file" "test" {
    name = "test"
    project = "default"
}