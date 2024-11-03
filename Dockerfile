ARG DOCKER_HUB_USERNAME=lijuthomas
ARG IMAGE_NAME=foodbuddy
ARG VERSION=latest

# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o foodbuddy ./cmd/main.go 

# Stage 2: Create a lightweight image for running the application
FROM alpine:latest

COPY .env ./

WORKDIR /root/

# Copy the built executable from the builder stage
COPY --from=builder /app/foodbuddy ./

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./foodbuddy"]
