#!/bin/bash

# Test Helm Chart with Minikube
# This script automates the process of testing the SMS Gateway Helm chart in a Minikube environment

# Exit on error
set -e

# Check for required commands
for cmd in minikube helm kubectl docker jq; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd command not found. Please install it first."
        exit 1
    fi
done

# Start Minikube if not running
if ! minikube status &> /dev/null; then
    echo "Starting Minikube..."
    minikube start --cpus=2 --memory=4096 --driver=docker
else
    echo "Minikube is already running"
fi

# Set Minikube Docker environment
echo "Setting up Minikube Docker environment..."
eval "$(minikube docker-env)"

# Create namespace for testing
NAMESPACE="sms-gateway-test"
kubectl create namespace $NAMESPACE || true

# Install the Helm chart
echo "Installing Helm chart..."
helm upgrade --install sms-gateway-test ./deployments/helm-chart \
  --namespace $NAMESPACE \
  --set image.pullPolicy=IfNotPresent \
  --set database.deployInternal=true \
  --set database.mariadb.rootPassword=root \
  --set database.password=sms \
  --set gateway.privateToken=test-token

# Wait for pods to be ready
echo "Waiting for pods to be ready..."
kubectl wait --namespace $NAMESPACE \
  --for condition=Available \
  deployment/sms-gateway-test-server \
  --timeout=120s

# Port forward to access the service
echo "Port forwarding to service (http://localhost:8080)..."
kubectl port-forward --namespace $NAMESPACE service/sms-gateway-test-server 8080:3000 &
PORT_FORWARD_PID=$!

# Wait for port-forward to be ready
echo "Waiting for port-forward to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "Port-forward is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: Port-forward failed to establish"
        kill $PORT_FORWARD_PID
        exit 1
    fi
    sleep 1
done

# Test the health endpoint
echo "Testing health endpoint..."
curl -s http://localhost:8080/health | jq . || echo "Health check failed"

# Run Helm tests
echo "Running Helm tests..."
helm test sms-gateway-test --namespace $NAMESPACE

# Cleanup
echo "Cleaning up..."
kill $PORT_FORWARD_PID
kubectl delete namespace $NAMESPACE

echo "Note: Minikube cluster is still running. Run 'minikube delete' to remove it."

echo -e "\n\033[32mTest completed successfully!\033[0m"
echo "You can inspect the test results above."