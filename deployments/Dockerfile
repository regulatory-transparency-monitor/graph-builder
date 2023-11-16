# syntax=docker/dockerfile:1
# A sample microservice in Go packaged into a container image.
FROM golang:1.19-alpine

# Set destination for COPY
WORKDIR /app

# Install curl
RUN apk --no-cache add curl

# Copy Go modules and dependencies to image
COPY go.mod go.sum ./
# download Go modules and dependencies
RUN go mod download
# Copy directory files, i.e all files ending with .go
COPY . . 

# RUN go get github.com/

# Build/ Compile applications
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /collector-service

# Build/Compile the application
RUN go build -o server ./cmd/

# Expose container on port 8080
EXPOSE 8080

# Copy the entrypoint script
COPY entrypoint.sh /app/

# Make the script executable
RUN chmod +x /app/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
