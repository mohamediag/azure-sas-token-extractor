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

# Use distroless as minimal base image to package the application
# https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/azure-sas-token-extractor .

# Set the entrypoint
ENTRYPOINT ["./azure-sas-token-extractor"]