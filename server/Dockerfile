FROM golang:1.20-alpine

WORKDIR /app

COPY ./server/go.mod ./server/go.sum ./

RUN go mod tidy

COPY ./server ./

RUN go build -o app main.go

EXPOSE 8080

CMD ["./app"]