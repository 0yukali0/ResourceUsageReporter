# Build stage
FROM ubuntu:26.04 AS builder

RUN apt-get update && apt-get install -y \
    golang-1.25 \
    build-essential \
    pkg-config \
    wget \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/lib/go-1.25/bin:$PATH"

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the executable with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o resource-reporter .

# Final stage - ubuntu image
FROM ubuntu:26.04

WORKDIR /root/

# Copy the binary from builder
COPY console.html .
COPY --from=builder /app/resource-reporter .

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./resource-reporter"]
