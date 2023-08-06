# Start your image with a node base image
FROM golang:1.19-alpine

# The /app directory should act as the main application directory
WORKDIR /app

COPY ./registry/registry.go /app/registry.go

RUN go build registry.go

ENTRYPOINT [ "./registry"]




