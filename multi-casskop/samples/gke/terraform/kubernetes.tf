###################################################
##        K8S Casskop install configuration      ##
## -                                             ##
###################################################

#################################
#       Helm : Tiller SA        #
#################################
// Create tiller service account on Kubernetes
resource "kubernetes_service_account" "tiller" {
  depends_on = ["google_container_node_pool.nodes"]
  metadata {
    name = "tiller"
    namespace = "${kubernetes_secret.tiller.metadata.0.namespace}"
  }
}
// Bind state of tiller's secret
resource "kubernetes_secret" "tiller" {
  metadata {
    name = "tiller"
    namespace = "kube-system"
  }
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
    name      = "${kubernetes_service_account.tiller.metadata.0.name}"
    namespace = "${kubernetes_service_account.tiller.metadata.0.namespace}"
  }
}

#################################
#          Namespaces           #
#################################
// Define prod namespace, for tracking pods deployment in production
resource "kubernetes_namespace" "cassandra-demo" {
  depends_on = ["google_container_node_pool.nodes"]
  metadata {
    annotations = {
      name = "cassandra-demo"
    }
    name = "cassandra-demo"
  }
}

// helm repository 
data "helm_repository" "casskop" {
  name = "casskop"
  url  = "https://Orange-OpenSource.github.io/cassandra-k8s-operator/helm"
}

// helm release
resource "helm_release" "casskop" {
  name             = "casskop"
  repository       = data.helm_repository.casskop.metadata[0].name
  chart            = "cassandra-operator"
  namespace        = kubernetes_namespace.cassandra-demo.metadata[0].name
  disable_webhooks = true
  #version          = "0.5.0-release"
  set {
    name  = "image.tag"
    value = "v0.5.0-release"
  }
  set {
    name  = "createCustomResource"
    value = true
  }
}