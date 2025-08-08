# Stage 1: Build the Go applications
FROM golang:1.24 AS builder

RUN apt-get update && apt-get install -y git

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Copy the entire backend source code
COPY . .

# Tidy and vendor dependencies
RUN go mod tidy

# Regenerate protobuf code
ENV PROTOC_VERSION=3.19.4
RUN apt-get update && apt-get install -y curl unzip && \
    curl -L https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip -o /tmp/protoc.zip && \
    unzip /tmp/protoc.zip -d /usr/local && \
    rm /tmp/protoc.zip
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
ENV PATH="$PATH:$(go env GOPATH)/bin"
RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/auth/proto/auth.proto

# Remove the conflicting resolver file
# RUN rm /app/services/gateway/resolver.go

# Build the main binary for each service
RUN go build -mod=vendor -o /app/gateway ./services/gateway
RUN go build -mod=vendor -o /app/auth ./services/auth
RUN go build -mod=vendor -o /app/user ./services/user
# Add build commands for other services here...

# Download wait-for-it.sh
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh

# Install grpc_health_probe
RUN go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.11

# Stage 2: Create the final image
FROM debian:stable-slim

RUN apt-get update && apt-get install -y bash

WORKDIR /app

# Copy the built binaries and wait-for-it.sh from the builder stage
COPY --from=builder /app/gateway /app/gateway
COPY --from=builder /app/auth /app/auth
COPY --from=builder /app/user /app/user
COPY --from=builder /usr/local/bin/wait-for-it.sh /usr/local/bin/wait-for-it.sh
COPY --from=builder /go/bin/grpc-health-probe /usr/local/bin/grpc-health-probe

# The command to run will be specified in the docker-compose.yml
# Example: CMD ["/app/gateway"]

# The command to run will be specified in the docker-compose.yml
# Example: CMD ["/app/gateway"]
