# Start your image with a node base image
FROM golang:1.20-alpine

# The /app directory should act as the main application directory
WORKDIR /app

COPY ./node/node.go /app/node.go
COPY ./node/sha_adapted.go /app/sha_adapted.go

RUN go build node.go sha_adapted.go

ENTRYPOINT [ "./node"]




