# Build stage
FROM golang:1.25-alpine AS build-env

# Install build dependencies
RUN apk add --no-cache make git libc-dev bash gcc linux-headers

WORKDIR /go/src/github.com/cosmos/example

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

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
