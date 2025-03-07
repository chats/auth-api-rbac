# Stage 1: Build the application
FROM golang:1.23.5-alpine AS builder

# Install git and other necessary packages
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth-api ./cmd/api

# Stage 2: Create the final image
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/auth-api .

# Copy configuration files
COPY config.yaml .

# Expose the port the API runs on
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/auth-api"]