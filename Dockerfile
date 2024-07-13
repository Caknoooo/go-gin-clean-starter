FROM golang:alpine

RUN apk update && apk upgrade && \
  apk add --no-cache bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8888

CMD ["./main"]