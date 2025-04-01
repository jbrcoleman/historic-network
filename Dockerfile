FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY *.go ./
COPY static/ ./static/

# Build the application with CGO disabled for better compatibility
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o historical-network-visualizer .

# Use a minimal alpine image for the final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests to Wikipedia
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/historical-network-visualizer .
COPY --from=builder /app/static ./static

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./historical-network-visualizer"]