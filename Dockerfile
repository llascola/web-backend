# --- Stage 1: Builder ---
FROM golang:1.25-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary named 'server'
RUN go build -o server main.go

# --- Stage 2: Runner ---
FROM alpine:latest

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the port
EXPOSE 8080

# Run the app
CMD ["./server"]
