# Use the official golang image as the base image
FROM golang:1.21 as builder

# Set working directory
WORKDIR /app


#copy go.mod and go.sum to the working directory
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy all files in the current directory to the /app directory in the container
COPY . /app

# Compile source code
RUN --mount=type=cache,target=/go/pkg/mod \
go mod download 

# Build the binary
RUN CGO_ENABLED=0  go build -ldflags "-s -w" -o /app/cdn main.go

# Install certificates (e.g., using `ca-certificates` package on Debian/Ubuntu-based images)
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates 



# Use the lightweight aslpine image as the basis for the final image
FROM scratch

# Set working directory
WORKDIR /root/

# Copy the compiled binaries from the build phase to the final image
COPY --from=builder /app/cdn /root/cdn

# Copy the certificate files to a known location
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#Specify the command to be executed when the container starts
ENTRYPOINT  ["/root/cdn"]