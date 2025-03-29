FROM golang:1.24 as base

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o go-app

EXPOSE 8084

CMD ["/build/go-app"]
