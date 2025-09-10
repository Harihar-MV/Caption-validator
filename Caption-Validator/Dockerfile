FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod file (if it exists) and source code
COPY . .

# Initialize go.mod
RUN go mod init caption-validator

# Download dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o caption-validator ./cmd/main.go

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install dependencies for file command
RUN apk --no-cache add file ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/caption-validator .

# Set entrypoint
ENTRYPOINT ["/app/caption-validator"]
