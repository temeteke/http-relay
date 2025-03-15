# Use the official Golang image as the base image
FROM golang:1.23.4

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o http-relay

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./http-relay"]