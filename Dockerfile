FROM golang:1.22 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o azure-sas-token-extractor

# Use a smaller base image for the final image
FROM alpine:latest

# Install CA certificates for making secure connections
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/azure-sas-token-extractor .

# Set the entrypoint
ENTRYPOINT ["./azure-sas-token-extractor"]