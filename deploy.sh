#!/bin/bash

# Set the namespace
NAMESPACE="foodbuddy"

# Delete all resources in the specified namespace
# kubectl delete all --all -n $NAMESPACE

# Build the Docker image
docker build -t lijuthomas/foodbuddy:latest .
docker push lijuthomas/foodbuddy:latest 

# Apply the namespace configuration
kubectl apply -f k8s/namespace.yaml

# Apply Kubernetes configurations
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/mysql-deployment.yaml
echo "Waiting for mysql to start..."
sleep 10  
kubectl apply -f k8s/app-deployment.yaml

# Wait for the pods to start (optional)
echo "Waiting for pods to start..."
sleep 10  

# Get the logs of the latest app pod
kubectl get pods -n $NAMESPACE

echo "performing port forwarding ...."
sleep 10 
kubectl port-forward -n foodbuddy service/foodbuddy 8080:80