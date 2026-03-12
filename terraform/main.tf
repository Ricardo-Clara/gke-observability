provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_container_cluster" "primary" {
  name     = "observability-cluster"
  location = var.region

  initial_node_count = 2

  node_config {
    machine_type = "e2-small"
    disk_type    = "pd-standard"
    disk_size_gb = 50

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}

output "kubeconfig_command" {
  value = "gcloud container clusters get-credentials ${google_container_cluster.primary.name} --region ${var.region} --project ${var.project_id}"
}
