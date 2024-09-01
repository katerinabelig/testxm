FROM golang:1.20-alpine

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bin/gocompany

CMD ["./bin/gocompany"]