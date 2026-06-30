# Kubernetes Go API Demo

A small go and Kubernetes REST API . The project is intentionally simple in order to use it as a clean portfolio project.

## Tech Stack

- Go standard library
- Docker
- Kubernetes

## API Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/health` | Returns service health information |
| `GET` | `/api/items` | Lists all items |
| `POST` | `/api/items` | Creates a new item |

## Run Locally

If Go is installed:

```bash
go run .
```

The API starts on port `8080` by default.

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/items
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Demo item","description":"Created from curl"}'
```

## Build With Docker

```bash
docker build -t kubernetes-go-api-demo .
docker run --rm -p 8080:8080 kubernetes-go-api-demo
```

## Deploy To Kubernetes With kind

Create a local Kubernetes cluster:

```bash
kind create cluster
```

Build the Docker image:

```bash
docker build -t kubernetes-go-api-demo .
```

Load the image into kind:

```bash
kind load docker-image kubernetes-go-api-demo
```

Apply the Kubernetes manifests:

```bash
kubectl apply -f k8s/
```

Useful Kubernetes commands:

```bash
kubectl get pods
kubectl logs <pod-name>
kubectl describe pod <pod-name>
kubectl port-forward service/kubernetes-go-api-demo 8080:80
```

After port forwarding, test the API:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/items
```

Create an item:

```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Kubernetes demo","description":"Running inside the cluster"}'
```

## Ingress

The Ingress manifest uses the host:

```text
kubernetes-go-api-demo.local
```

If you use an ingress controller locally, add this entry to your hosts file:

```text
127.0.0.1 kubernetes-go-api-demo.local
```

Then open:

```text
http://kubernetes-go-api-demo.local/health
```

## Project Structure

```text
.
├── Dockerfile
├── README.md
├── go.mod
├── main.go
├── postman_collection.json
└── k8s
    ├── configmap.yaml
    ├── deployment.yaml
    ├── ingress.yaml
    └── service.yaml
```
