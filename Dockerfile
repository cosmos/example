# Build stage
FROM golang:1.26-alpine AS build-env

# Install build dependencies
RUN apk add --no-cache make git libc-dev bash gcc linux-headers

WORKDIR /go/src/github.com/cosmos/example

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

# Declared without defaults so BuildKit supplies the build platform. Hardcoding a
# GOARCH here cross-compiles for that architecture regardless of the host, and the
# resulting binary then runs under emulation. Emulated x86 mis-executes the AVX2
# chacha20poly1305 assembly, which breaks the CometBFT P2P handshake and leaves
# localnet nodes unable to peer with each other.
ARG TARGETOS
ARG TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /go/bin/exampled ./exampled

FROM alpine:3

RUN apk add --no-cache bash jq sed curl

COPY --from=build-env /go/bin/exampled /usr/bin/exampled

EXPOSE 26656 26657 1317 9090

WORKDIR /root

COPY scripts/localnet/wrapper.sh /usr/bin/wrapper.sh
RUN chmod +x /usr/bin/wrapper.sh

ENTRYPOINT []
CMD ["exampled", "start"]
