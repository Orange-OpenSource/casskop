// Documentations support : https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/gke.md
// Manage DNS Zone for external dns.
resource "google_dns_managed_zone" "external-dns-zone" {
    count = var.create_dns ? 1 : 0

    name         = "external-dns-test-gcp-trycatchlearn-fr"
    dns_name     = "external-dns-test.gcp.trycatchlearn.fr."
    description  = "Automatically managed zone by kubernetes.io/external-dns"
    labels       = {}
    visibility   = "public"
}

// Manage DNS record sets for external dns zone.
resource "google_dns_record_set" "external-dns-record-set" {
    count = var.create_dns ? 1 : 0
    managed_zone = "tracking-pdb"
    name         = "external-dns-test.gcp.trycatchlearn.fr."
    project      = "poc-rtc"
    rrdatas      = google_dns_managed_zone.external-dns-zone[0].name_servers
    ttl          = 300
    type         = "NS"
}


// Cluster role with permissions required for external dns.
resource "kubernetes_cluster_role" "external-dns" {
  metadata {
    name = "external-dns"
  }

  rule {
    api_groups = ["",]
    resources  = ["services",]
    verbs       = ["get", "watch", "list",]
  }

  rule {
    api_groups = ["",]
    resources  = ["pods",]
    verbs       = ["get", "watch", "list",]
  }

  rule {
    api_groups = ["",]
    resources  = ["nodes",]
    verbs       = ["get", "watch", "list",]
  }

  rule {
      api_groups = ["extensions",]
      resources  = ["ingresses",]
      verbs      = ["get", "watch", "list",]
    }
}

// Create service account for external-dns
resource "kubernetes_service_account" "external-dns" {
  depends_on = ["google_container_node_pool.nodes"]
  metadata {
    name = "external-dns"
    namespace = "${kubernetes_namespace.cassandra-demo.metadata.0.name}"
  }
  automount_service_account_token = true
}

// Binding external-dns cluster role, with the external-dns Service account.
resource "kubernetes_cluster_role_binding" "external-dns-viewer" {
  metadata {
    name      = "external-dns-viewer"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "${kubernetes_cluster_role.external-dns.metadata.0.name}"
  }
  subject {
    kind      = "ServiceAccount"
    name      = "${kubernetes_service_account.external-dns.metadata.0.name}"
    namespace = "${kubernetes_service_account.external-dns.metadata.0.namespace}"
  }
}

// Create deployment 
resource "kubernetes_deployment" "external-dns" {
  metadata {
    annotations      = {}
    labels           = {
      "app" = "external-dns"
    }
    name             = "external-dns"
    namespace = "${kubernetes_service_account.external-dns.metadata.0.namespace}"
  }

  spec {
    selector {
      match_labels = {
        "app" = "external-dns"
      }
    }

    strategy {
      type = "Recreate"
    }

    template {
      metadata {
        annotations = {}
        labels      = {
          "app" = "external-dns"
        }
      }

      spec {

        service_account_name = kubernetes_service_account.external-dns.metadata.0.name
        automount_service_account_token = true
        container {
          args = [
            "--source=service",
            "--source=ingress",
            "--domain-filter=external-dns-test.gcp.trycatchlearn.fr",
            "--provider=google",
            "--policy=upsert-only",
            "--registry=txt",
            "--txt-owner-id=gke-${google_container_cluster.cassandra-cluster.project}-${google_container_cluster.cassandra-cluster.location}_${google_container_cluster.cassandra-cluster.name}-${kubernetes_namespace.cassandra-demo.metadata[0].name}",
          ]
          command                  = []
          image                    = "registry.opensource.zalan.do/teapot/external-dns:latest"
          name                     = "external-dns"
        }
      }
    }
  }
}
