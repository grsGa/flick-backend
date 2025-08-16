# Stage 1: Build the Go applications
FROM golang:1.24 AS builder

RUN apt-get update && apt-get install -y git

WORKDIR /app

# Copy the entire backend source code (including go.work and all service go.mod files)
COPY . .

# Set environment variables to allow external deps but force local resolution for our modules
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOPRIVATE=github.com/flick/backend
ENV GO111MODULE=on

# Use go work mode to build all services independently
# Each service has its own go.mod and dependencies are managed through go.work

# Initialize workspace and download external dependencies first
RUN go work use -r .
RUN go work sync

# Pre-download external dependencies for all services to avoid network issues during build
WORKDIR /app/services/gateway
RUN go mod download

WORKDIR /app/services/auth  
RUN go mod download

WORKDIR /app/services/user
RUN go mod download

WORKDIR /app/services/media
RUN go mod download

WORKDIR /app/services/content
RUN go mod download

WORKDIR /app/services/interaction
RUN go mod download

WORKDIR /app/services/messages
RUN go mod download

WORKDIR /app/services/notification
RUN go mod download

WORKDIR /app/services/recommendation
RUN go mod download

WORKDIR /app/services/search
RUN go mod download

WORKDIR /app

# Build each service independently using their own go.mod
WORKDIR /app/services/gateway
RUN go build -o /app/gateway .

WORKDIR /app/services/auth
RUN go build -o /app/auth .

WORKDIR /app/services/user
RUN go build -o /app/user .

WORKDIR /app/services/media
RUN go build -o /app/media .

WORKDIR /app/services/content
RUN go build -o /app/content .

WORKDIR /app/services/interaction
RUN go build -o /app/interaction .

WORKDIR /app/services/messages
RUN go build -o /app/messages .

WORKDIR /app/services/notification
RUN go build -o /app/notification .

WORKDIR /app/services/recommendation
RUN go build -o /app/recommendation .

WORKDIR /app/services/search
RUN go build -o /app/search .

WORKDIR /app

# Download wait-for-it.sh
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh

# Install grpc_health_probe
RUN go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.11

# Stage 2: Create the final image
FROM debian:stable-slim

RUN apt-get update && apt-get install -y bash ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built binaries and wait-for-it.sh from the builder stage
COPY --from=builder /app/gateway /app/gateway
COPY --from=builder /app/auth /app/auth
COPY --from=builder /app/user /app/user
COPY --from=builder /app/media /app/media
COPY --from=builder /app/content /app/content
COPY --from=builder /app/interaction /app/interaction
COPY --from=builder /app/messages /app/messages
COPY --from=builder /app/notification /app/notification
COPY --from=builder /app/recommendation /app/recommendation
COPY --from=builder /app/search /app/search
COPY --from=builder /usr/local/bin/wait-for-it.sh /usr/local/bin/wait-for-it.sh
COPY --from=builder /go/bin/grpc-health-probe /usr/local/bin/grpc-health-probe

# The command to run will be specified in the docker-compose.yml
# Example: CMD ["/app/gateway"]
