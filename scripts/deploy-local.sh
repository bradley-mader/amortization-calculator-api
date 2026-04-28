#!/usr/bin/env bash
set -euo pipefail

# Start minikube if not running
if ! minikube status | grep -q "Running"; then
    echo "Starting minikube..."
    minikube start
fi

# Use minikube context
kubectl config use-context minikube

# Build image in minikube's Docker daemon
eval $(minikube docker-env)
docker build -t amortization-api:local .

# Apply local-dev overlay
kubectl apply -k k8s/overlays/local-dev

# Wait for rollout
kubectl rollout status deployment/amortization-api --timeout=60s

echo ""
echo "Deployment successful!"
echo "To access the service, run:"
echo "  kubectl port-forward svc/amortization-api 8080:8080"
