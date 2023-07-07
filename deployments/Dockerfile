# syntax=docker/dockerfile:1
# A sample microservice in Go packaged into a container image.
FROM golang:1.19-alpine

# Set destination for COPY
WORKDIR /app

# Copy Go modules and dependencies to image
COPY go.mod go.sum ./
# download Go modules and dependencies
RUN go mod download
# Copy directory files, i.e all files ending with .go
COPY . . 

# RUN go get github.com/

# Build/ Compile applications
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /collector-service

# compile application
RUN go build -o server .

# Expose container on port 8080
EXPOSE 8080

# ENTRYPOINT CompileDaemon --build="go build cmd/api/main.go" --command=./main 

# Run
CMD ["server"]