FROM golang:1.19-bullseye AS builder

COPY . /src

WORKDIR /src

RUN go build -o ./build/depviz ./cmd/depviz/main.go

FROM ubuntu:20.04

RUN apt-get -y update \
    && apt-get -y install ca-certificates

COPY --from=builder /src/build/depviz /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/depviz"]
