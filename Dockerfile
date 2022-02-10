FROM golang:1.17.6-alpine
MAINTAINER Alexandre Ferland <me@alexferl.com>

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build
RUN mv uberwachen /uberwachen

ENTRYPOINT ["/uberwachen"]

EXPOSE 1323
CMD []
