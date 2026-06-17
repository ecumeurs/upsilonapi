# Base image
FROM golang:1.25-alpine

# Install build dependencies
RUN apk add --no-cache build-base curl

# Set working directory to root to maintain workspace structure
WORKDIR /app

# Copy go.work and module files for caching
COPY go.work go.work.sum ./
COPY upsilonapi/go.mod upsilonapi/go.su[m] ./upsilonapi/
COPY upsilonbattle/go.mod upsilonbattle/go.su[m] ./upsilonbattle/
COPY upsiloncli/go.mod upsiloncli/go.su[m] ./upsiloncli/
COPY upsilonmapdata/go.mod upsilonmapdata/go.su[m] ./upsilonmapdata/
COPY upsilonmapmaker/go.mod upsilonmapmaker/go.su[m] ./upsilonmapmaker/
COPY upsilonserializer/go.mod upsilonserializer/go.su[m] ./upsilonserializer/
COPY upsilontools/go.mod upsilontools/go.su[m] ./upsilontools/
COPY upsilontypes/go.mod upsilontypes/go.su[m] ./upsilontypes/

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the battle engine API
WORKDIR /app/upsilonapi
RUN go build -o /app/upsilon-engine .

# Final execution
CMD ["/app/upsilon-engine"]
