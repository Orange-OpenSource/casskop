###################################################
##        K8S Casskop install configuration      ##
## -                                             ##
###################################################

#################################
#       Helm : Tiller SA        #
#################################
// Bind state of tiller's secret
resource "kubernetes_secret" "tiller" {
  metadata {
    name = "tiller"
    namespace = "kube-system"
  }
}

// Create tiller service account on Kubernetes
resource "kubernetes_service_account" "tiller" {
  metadata {
    name = "tiller"
    namespace = kubernetes_secret.tiller.metadata.0.namespace
  }  
  depends_on = [google_container_node_pool.nodes]
}

// Bind tiller service account with cluster role admin on K8S
resource "kubernetes_cluster_role_binding" "tiller-admin-binding" {
  metadata {
    name      = "tiller-admin-binding"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.tiller.metadata.0.name
    namespace = kubernetes_service_account.tiller.metadata.0.namespace
  }
  depends_on = [kubernetes_service_account.tiller]
}

#################################
#          Namespaces           #
#################################
// Define prod namespace, for tracking pods deployment in production
resource "kubernetes_namespace" "cassandra-demo" {
  metadata {
    annotations = {
      name = var.namespace
    }
    name = var.namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

// Config map cassandra.
resource "kubernetes_config_map" "cassandra" {
    data        = {
        "post_run.sh" = file("${path.module}/post_run.sh") 
        "pre_run.sh"  = file("${path.module}/pre_run.sh")
    }

    metadata {
        annotations  = {}
        labels       = {}
        name         = "cassandra-configmap-v1"
        namespace    = kubernetes_namespace.cassandra-demo.metadata[0].name
    }
}

// Storage class
resource "kubernetes_storage_class" "cassandra-standard" {
  parameters          = {
      "type" = "pd-standard"
  }
  reclaim_policy      = "Delete"
  storage_provisioner = "kubernetes.io/gce-pd"
  volume_binding_mode = "WaitForFirstConsumer"

  metadata {
    annotations      = {}
    labels           = {}
    name             = "standard-wait"
  }
  depends_on = [google_container_node_pool.nodes]
}

// helm repository 
data "helm_repository" "casskop" {
  name = "casskop"
  url  = "https://Orange-OpenSource.github.io/cassandra-k8s-operator/helm"

  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
}

// helm release
resource "helm_release" "casskop" {
  name             = "casskop"
  repository       = data.helm_repository.casskop.metadata[0].name
  chart            = "cassandra-operator"
  namespace        = kubernetes_namespace.cassandra-demo.metadata[0].name
  disable_webhooks = false
  #version          = "0.5.0-release"
  set {
    name  = "image.tag"
    value = "v0.5.0-release"
  }
  set {
    name  = "createCustomResource"
    value = true
  }
  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
}