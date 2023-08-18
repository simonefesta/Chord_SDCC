# Start your image with a node base image
FROM golang:1.20-alpine

# The /app directory should act as the main application directory
WORKDIR /app

COPY ./config.json /app/config.json 
COPY ./registry/registry.go /app/registry.go
COPY ./registry/utils.go /app/utils.go


EXPOSE 1234

RUN go build registry.go utils.go

ENTRYPOINT [ "./registry"]




