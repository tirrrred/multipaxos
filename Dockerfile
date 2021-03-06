#Multistage builder
FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /dummyGo

# Copy and download dependency using go mod
COPY go.mod .
#COPY go.sum .
#RUN go mod download

# Copy the code into the container
COPY . .

RUN apk update && apk add nano
# Build the application
#RUN go build -o main .

# Move to /dist directory as the place for resulting binary folder
#WORKDIR /dist

# Copy binary from build to main folder
#RUN cp /build/main .

# Build a small image
#FROM scratch

#Copy static files
#COPY ./lab5/netConf.json ./lab5/netConf.json

#COPY --from=builder /dist/main /

# Command to run
#ENTRYPOINT ["/main"]