# Use the official Go image
FROM golang:1.24 as builder

# Set working directory inside the container
WORKDIR /app

# Copy the source code
COPY . .

# Download dependencies and build the binary
RUN go mod tidy && go build -o dropgox-backend main.go

# Use a smaller base image for the final container
FROM alpine:latest

# Set working directory in the smaller container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=builder /app/dropgox-backend .

# Expose the application port
EXPOSE 9090

# Run the application
CMD ["./dropgox-backend"]
