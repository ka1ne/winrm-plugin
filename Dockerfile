# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for dependency management
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o winrm-plugin .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S winrm && \
    adduser -u 1001 -S winrm -G winrm

# Set working directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/winrm-plugin .

# Change ownership to non-root user
RUN chown winrm:winrm winrm-plugin

# Switch to non-root user
USER winrm

# Set entrypoint
ENTRYPOINT ["./winrm-plugin"] 