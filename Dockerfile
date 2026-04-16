# Base image
FROM golang:1.25-alpine

# Install build dependencies
RUN apk add --no-color build-base

# Set working directory to root to maintain workspace structure
WORKDIR /app

# Copy go.work and module files for caching
COPY go.work go.work.sum ./
COPY upsilonapi/go.mod upsilonapi/go.sum ./upsilonapi/
COPY upsilonbattle/go.mod upsilonbattle/go.sum ./upsilonbattle/
COPY upsilonmapdata/go.mod upsilonmapdata/go.sum ./upsilonmapdata/
COPY upsilonmapmaker/go.mod upsilonmapmaker/go.sum ./upsilonmapmaker/
COPY upsilonserializer/go.mod upsilonserializer/go.sum ./upsilonserializer/
COPY upsilontools/go.mod upsilontools/go.sum ./upsilontools/

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the battle engine
WORKDIR /app/upsilonbattle
RUN go build -o /app/upsilon-engine .

# Final execution
CMD ["/app/upsilon-engine"]
