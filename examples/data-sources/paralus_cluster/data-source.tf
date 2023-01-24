# This example shows a cluster data source

data "paralus_cluster" "test" {
    name = "test"
    project = "default"
}