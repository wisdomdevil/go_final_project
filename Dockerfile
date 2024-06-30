FROM golang:1.22.2

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /service cmd/main.go

EXPOSE 7540

CMD ["/service"]