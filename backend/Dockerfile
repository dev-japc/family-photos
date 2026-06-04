FROM golang:1.25.4-alpine

RUN apk add --no-cache git build-base gcc musl-dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


# Comando por defecto para desarrollo usando air
CMD ["air"]
