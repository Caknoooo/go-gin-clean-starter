FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache git
RUN go install github.com/air-verse/air@v1.65.1

COPY . .

RUN go mod tidy

CMD ["air"]
