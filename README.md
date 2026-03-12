# gke-observability

A Go API deployed on Google Kubernetes Engine (GKE), provisioned with Terraform, and monitored with Prometheus + Grafana. CI/CD is handled via GitHub Actions, which also queries live metrics after every deployment.

---

## Stack

| Layer | Technology |
| --- | --- |
| Language | Go |
| Container Runtime | Docker |
| Orchestration | Kubernetes (GKE) |
| Infrastructure | Terraform |
| Monitoring | Prometheus + Grafana (kube-prometheus-stack) |
| CI/CD | GitHub Actions |
| Cloud | Google Cloud Platform (GCP) |

---

## Architecture

```text
GitHub Actions
     │
     ├── Build & push Docker image ──► GCP Artifact Registry
     │
     ├── Terraform ──► GKE Cluster
     │
     ├── kubectl apply ──► go-api Deployment + Service
     │
     ├── Smoke test ──► generate HTTP traffic
     │
     └── Query Prometheus ──► display metrics in workflow summary

Cluster (monitoring namespace)
     ├── Prometheus ──► scrapes all pods with prometheus.io/scrape: "true"
     └── Grafana ──────► visualizes Prometheus data
```

---

## Project Structure

```text
gke-observability/
├── terraform/
│   ├── main.tf           # GKE cluster definition
│   ├── variables.tf      # project_id, region
│   └── outputs.tf        # kubeconfig command
├── app/
│   ├── main.go           # Go API with /health and /metrics endpoints
│   └── Dockerfile        # Multi-stage build
├── k8s/
│   ├── deployment.yaml   # go-api Deployment with Prometheus annotations
│   └── service.yaml      # LoadBalancer Service
├── .github/
│   └── workflows/
│       └── deploy.yml    # Full CI/CD pipeline
└── README.md
```

---

## Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/) >= 3.0
- [Docker](https://docs.docker.com/get-docker/)
- A GCP project with billing enabled

---

## Setup

### 1. Clone the repo

```bash
git clone https://github.com/Ricardo-Clara/gke-observability
cd gke-observability
```

### 2. Provision infrastructure with Terraform

```bash
cd terraform
terraform init
terraform apply -var="project_id=YOUR_PROJECT_ID"
```

This creates a GKE cluster with 2 `e2-small` nodes in `europe-west1`.

### 3. Configure kubectl

```bash
gcloud container clusters get-credentials observability-cluster \
  --region europe-west1 \
  --project YOUR_PROJECT_ID
```

### 4. Build and push the Docker image

```bash
gcloud auth configure-docker europe-west1-docker.pkg.dev

docker build -t europe-west1-docker.pkg.dev/YOUR_PROJECT/api/go-api:v1 ./app
docker push europe-west1-docker.pkg.dev/YOUR_PROJECT/api/go-api:v1
```

### 5. Deploy the app

```bash
kubectl apply -f k8s/
kubectl get pods      # wait until Running
kubectl get svc       # copy EXTERNAL-IP
```

### 6. Deploy Prometheus + Grafana

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin123
```

---

## Accessing Monitoring

### Grafana

```bash
kubectl port-forward svc/monitoring-grafana 3000:80 -n monitoring
```

Open [http://localhost:3000](http://localhost:3000) — login: `admin / admin123`

- Import dashboard ID `6417` for Kubernetes app metrics
- Query `http_requests_total` in the Explore tab

### Prometheus

```bash
kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090:9090 -n monitoring
```

Open [http://localhost:9090](http://localhost:9090)

- Check **Status → Targets** to confirm `go-api` is being scraped
- Example query: `rate(http_requests_total[5m])`

---

## How Prometheus Discovers the App

Prometheus automatically scrapes any pod annotated with:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

No additional configuration needed. Add these annotations to any new app and it will be monitored automatically.

---

## CI/CD Pipeline

Every push to `main`:

1. Builds and pushes a new Docker image tagged with the commit SHA
2. Deploys to GKE via `kubectl set image`
3. Waits for rollout to complete
4. Hits the `/health` endpoint 20 times to generate traffic
5. Waits 35s for Prometheus to scrape the new metrics
6. Queries Prometheus and prints a metrics summary to the workflow summary page

### Required GitHub Secrets

| Secret | Description |
| --- | --- |
| `GCP_SA_KEY` | JSON key for a GCP Service Account with GKE + Artifact Registry permissions |
| `GCP_PROJECT` | Your GCP project ID |

---

## API Endpoints

| Endpoint | Description |
| --- | --- |
| `GET /health` | Returns `{"status":"ok"}` — also increments request counter |
| `GET /metrics` | Prometheus metrics scrape endpoint |

---

## Tearing Down

```bash
cd terraform
terraform destroy -var="project_id=YOUR_PROJECT_ID"
```
