# This example shows how to import a cluster into paralus

resource "paralus_cluster" "testcluster" {
    name = "clusterresource"
    project = "test"
    cluster_type = "imported"
    params {
        provision_type = "IMPORT"
        provision_environment = "CLOUD"
        kubernetes_provider = "EKS"
        state = "PROVISION"
    }
}