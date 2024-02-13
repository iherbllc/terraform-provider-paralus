# This example shows a kubeconfig file data source

data "paralus_kubeconfig" "test" {
    name = "test@someplace.com"
    cluster = "default"
}