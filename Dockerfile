FROM golang:1.9
COPY . /go/src/mundial-go-ms
WORKDIR /go/src/mundial-go-ms
RUN go install -ldflags="-s -w" ./cmd/...
