# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

WORKDIR /root

# Install runtime dependencies
RUN apk add --no-cache \
    mysql-client \
    postgresql-client \
    curl \
    bash

# Copy binary from builder
COPY --from=builder /app/build/bakdb /usr/local/bin/bakdb

# Create config directory
RUN mkdir -p /root/.bakdb

# Expose entry point
ENTRYPOINT ["bakdb"]
