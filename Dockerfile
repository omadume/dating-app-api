# Use the official Golang image as the base image
FROM golang:1.22.4

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Download Go module dependencies
RUN go mod tidy

# Build the Go app
RUN go build -o main .

# Command to run the executable
CMD ["./main"]
