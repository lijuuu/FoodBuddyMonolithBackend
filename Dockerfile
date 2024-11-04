# Arguments for Docker Hub and image version
ARG DOCKER_HUB_USERNAME=lijuthomas
ARG IMAGE_NAME=foodbuddy
ARG VERSION=latest

# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o foodbuddy ./cmd/main.go 

# Stage 2: Create the final lightweight image
FROM alpine:latest
WORKDIR /root/
COPY .env ./
COPY --from=builder /app/foodbuddy ./
EXPOSE 8080
CMD ["./foodbuddy"]