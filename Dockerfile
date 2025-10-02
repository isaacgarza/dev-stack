# Build stage
FROM golang:1.24-alpine AS builder

# Install git for fetching dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o dev-stack ./cmd/dev-stack

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and docker client
RUN apk --no-cache add ca-certificates docker-cli docker-cli-compose

# Create non-root user
RUN addgroup -g 1001 -S devstack && \
    adduser -u 1001 -S devstack -G devstack

# Set working directory
WORKDIR /home/devstack

# Copy the binary from builder stage
COPY --from=builder /app/dev-stack /usr/local/bin/dev-stack

# Change ownership and make executable
RUN chown devstack:devstack /usr/local/bin/dev-stack && \
    chmod +x /usr/local/bin/dev-stack

# Switch to non-root user
USER devstack

# Expose any ports if needed (optional)
# EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["dev-stack"]

# Default command
CMD ["--help"]
