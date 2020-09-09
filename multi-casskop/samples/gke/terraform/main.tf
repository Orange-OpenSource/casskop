// VARIABLES
variable "namespace" {  
  description = ""
  type        = string
}

variable "service_account_json_file" {
  description = "Path to local service account's json file"
  type        = string
}

variable "cluster_region" {
  description = ""
  type        = string
}
variable "cluster_zone" {
  description = ""
  type        = string
}

variable "username" {
  description = ""
  type        = string
}

variable "password" {
  description = ""
  type        = string
}

variable "project" {
  description = "GCP project name"
  type        = string
}

variable "master_state" {
  description = "Can be slave or master, it is only an additional tag for cluster name"
  type        = string
  default     = "slave"
}

variable "helm_version" {
  description = "If specified set the helm version used"
  type        = string
  default     = "v2.15.1"
}

variable "create_dns" {
  description = "If set to 1, create cloud dns for external-DNS"
  type        = bool
}

variable "dns_zone_name" {
  description = ""
  type        = string
}

variable "dns_name" {
  description = ""
  type        = string
}

variable "managed_zone" {
  description = ""
  type        = string
}

variable "casskop_image_tag" {
  description = ""
  type        = string
  default     = "v0.5.5-release"
}

// Provider definition
provider "google" {
  project = var.project
  credentials = file(var.service_account_json_file)
  region  = var.cluster_region
  zone    = var.cluster_zone
}

// Provider definition for beta features
provider "google-beta" {
  project = var.project
  credentials = file(var.service_account_json_file)
  region  = var.cluster_region
  zone    = var.cluster_zone
}

// Define K8S provider
provider "kubernetes" {
    load_config_file       = false
    host                   = google_container_cluster.cassandra-cluster.endpoint
    username               = google_container_cluster.cassandra-cluster.master_auth.0.username
    password               = google_container_cluster.cassandra-cluster.master_auth.0.password
    client_certificate     = base64decode(google_container_cluster.cassandra-cluster.master_auth.0.client_certificate)
    client_key             = base64decode(google_container_cluster.cassandra-cluster.master_auth.0.client_key)
    cluster_ca_certificate = base64decode(google_container_cluster.cassandra-cluster.master_auth.0.cluster_ca_certificate)
}

// Define Helm provider
provider "helm" {
    install_tiller  = true
    tiller_image    = "gcr.io/kubernetes-helm/tiller:${var.helm_version}"
    service_account = kubernetes_service_account.tiller.metadata.0.name
    debug           = true
    
    kubernetes {
        host                   = google_container_cluster.cassandra-cluster.endpoint
        token                  = data.google_client_config.current.access_token
        client_certificate     = base64decode(google_container_cluster.cassandra-cluster.master_auth.0.client_certificate)
        client_key             = base64decode(google_container_cluster.cassandra-cluster.master_auth.0.client_key)
        cluster_ca_certificate = base64decode(google_container_cluster.cassandra-cluster.master_auth.0.cluster_ca_certificate)
    }
}
