###################################################
##          GKE node pools configuration         ##
## - for global services (ex: Istio)             ##
###################################################
resource "google_container_node_pool" "nodes" {
    provider = google-beta

    name = "nodes-pool"
    name_prefix = ""
    location = var.cluster_zone
    cluster = google_container_cluster.cassandra-cluster.name

    # he number of nodes per instance group. This field can be used to update the number of nodes per instance group but should not be used alongside autoscaling
    node_count = 4

    # The maximum number of pods per node in this node pool. Note that this does not work on node pools which are "route-based" - that is, node pools belonging to clusters that do not have IP Aliasing enabled
    max_pods_per_node = 0

    #  Node management configuration, wherein auto-repair and auto-upgrade is configured
    management {
        auto_repair = true
        auto_upgrade = true
    }

    
    # The node configuration of the pool
    node_config {

        #Minimum CPU platform to be used by this instance. The instance may be scheduled on the specified or newer CPU platform. Applicable values are the friendly names of CPU platforms, such as Intel Haswell
        min_cpu_platform = ""

        # Size of the disk attached to each node, specified in GB. The smallest allowed disk size is 10GB. Defaults to 100GB
        disk_size_gb = 100

        # Type of the disk attached to each node (e.g. 'pd-standard' or 'pd-ssd'). If unspecified, the default disk type is 'pd-standard'
        disk_type = "pd-standard"

        #  List of the type and count of accelerator cards attached to the instance.
        ## type : The accelerator type resource to expose to this instance. E.g. nvidia-tesla-k80.
        ## count : The number of the guest accelerator cards exposed to this instance
        guest_accelerator = []

        # The image type to use for this node. Note that changing the image type will delete and recreate all nodes in the node pool
        image_type = "COS"

        #The Kubernetes labels (key/value pairs) to be applied to each node
        labels = {}

        # The amount of local SSD disks that will be attached to each cluster node
        local_ssd_count = 0

        # The name of a Google Compute Engine machine type. Defaults to n1-standard-1
        machine_type = "n1-standard-2"

        # A boolean that represents whether or not the underlying node VMs are preemptible
        preemptible  = false

        # The metadata key/value pairs assigned to instances in the cluster. From GKE 1.12 onwards, disable-legacy-endpoints is set to true by the API
        metadata = {
            disable-legacy-endpoints = "true"
        }

        # The set of Google API scopes to be made available on all of the node VMs under the "default" service account. These can be either FQDNs, or scope aliases. The following scopes are necessary to ensure the correct functioning of the cluster
        #  - storage-ro (https://www.googleapis.com/auth/devstorage.read_only), if the cluster must read private images from GCR. Note this will grant read access to ALL GCS content unless you also specify a custom role
        #  - logging-write (https://www.googleapis.com/auth/logging.write), if logging_service points to Google 
        #  - monitoring (https://www.googleapis.com/auth/monitoring), if monitoring_service points to Google 
        oauth_scopes = [
            "https://www.googleapis.com/auth/devstorage.read_only",
            "https://www.googleapis.com/auth/logging.write",
            "https://www.googleapis.com/auth/monitoring",
            "https://www.googleapis.com/auth/service.management.readonly",
            "https://www.googleapis.com/auth/servicecontrol",
            "https://www.googleapis.com/auth/trace.append",
            "https://www.googleapis.com/auth/ndev.clouddns.readwrite" # Cloud DNS access scope.
        ]

        # The service account to be used by the Node VMs. If not specified, the "default" service account is used. In order to use the configured oauth_scopes for logging and monitoring, the service account being used needs the roles/logging.logWriter and roles/monitoring.metricWriter roles
        service_account = "default"

        # The list of instance tags applied to all nodes. Tags are used to identify valid sources or targets for network firewalls.
        tags = ["cassandra-cluster"]
    }
}

resource "google_compute_firewall" "cassandra-cluster" {
    count = var.create_dns ? 1 : 0

    direction               = "INGRESS"
    disabled                = false
    name                    = "gke-cassandra-cluster"
    network                 = "https://www.googleapis.com/compute/v1/projects/poc-rtc/global/networks/default"
    priority                = 1000
    source_ranges           = [
        "0.0.0.0/0",
    ]
    target_tags             = [
        "cassandra-cluster",
    ]

    allow {
        ports    = []
        protocol = "tcp"
    }
    allow {
        ports    = []
        protocol = "udp"
    }

    depends_on = [google_container_node_pool.nodes]
}
