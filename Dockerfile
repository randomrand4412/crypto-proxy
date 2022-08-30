# hadolint ignore=DL3007
FROM golang:1.18.4
WORKDIR /go/src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy cmd/crypto-proxy/main.go

FROM alpine:latest  
WORKDIR /root/
COPY --from=0 /go/src/proxy .
ENTRYPOINT ["./proxy"]  
