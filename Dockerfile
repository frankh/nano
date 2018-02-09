FROM golang:1.9 AS gobuild

WORKDIR /go/src/github.com/frankh/nano
RUN go get \
  github.com/frankh/crypto/ed25519 \
  github.com/golang/crypto/blake2b \
  github.com/pkg/errors \
  github.com/dgraph-io/badger

COPY . ./

RUN go build -o nano .


FROM debian:8-slim

COPY --from=gobuild /go/src/github.com/frankh/nano/nano /nano

ENTRYPOINT ["/nano"]
