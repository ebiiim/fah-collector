FROM golang:1.16-buster as builder
WORKDIR /go/src/app
COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 go build "-ldflags=-s -w" -trimpath -o main

FROM alpine:3.13
LABEL org.opencontainers.image.source https://github.com/ebiiim/fah-collector
COPY --from=builder /go/src/app/main .
ENTRYPOINT [ "./main" ]
