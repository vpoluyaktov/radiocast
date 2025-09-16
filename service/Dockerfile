# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy only necessary source files for compilation
COPY *.go ./
COPY internal/ ./internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o radiocast .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and ffmpeg + fonts for GIF generation
RUN apk --no-cache add ca-certificates tzdata ffmpeg ttf-dejavu

# Create non-root user first
RUN adduser -D -s /bin/sh radiocast

WORKDIR /home/radiocast

# Copy only the binary and essential runtime files from builder stage
COPY --from=builder --chown=radiocast:radiocast /app/radiocast .
COPY --from=builder --chown=radiocast:radiocast /app/internal/templates ./internal/templates
COPY --from=builder --chown=radiocast:radiocast /app/internal/assets ./internal/assets

# Make binary executable
RUN chmod +x radiocast

USER radiocast

# Expose port
EXPOSE 8981

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8981/health || exit 1

# Run the application
CMD ["./radiocast", "-deployment=gcs"]
