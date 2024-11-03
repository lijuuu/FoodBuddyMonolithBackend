#!/bin/bash

# Set the namespace
NAMESPACE="foodbuddy"

# Delete all resources in the specified namespace
kubectl delete all --all -n $NAMESPACE

# Apply the namespace configuration
kubectl apply -f k8s/namespace.yaml

# Build the Docker image
docker build -t lijuthomas/foodbuddy:latest .
docker push lijuthomas/foodbuddy:latest 


# Apply Kubernetes configurations
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/mysql-deployment.yaml
kubectl apply -f k8s/app-deployment.yaml

# Wait for the pods to start (optional)
echo "Waiting for pods to start..."
sleep 10  

# Get the logs of the latest app pod
kubectl get pods -n $NAMESPACE
