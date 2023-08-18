# Start your image with a node base image
FROM golang:1.20-alpine

# The /app directory should act as the main application directory
WORKDIR /app

COPY ./config.json /app/config.json 
COPY ./node/node.go /app/node.go
COPY ./node/utils.go /app/utils.go

RUN go build node.go utils.go

ENTRYPOINT [ "./node"]




