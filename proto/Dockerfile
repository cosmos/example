# This Dockerfile is used for proto generation
# To build, run `make proto-image-build`

FROM bufbuild/buf:1.7.0 as BUILDER

FROM golang:1.23-alpine


RUN apk add --no-cache \
  nodejs \
  npm \
  git \
  make

ENV GOLANG_PROTOBUF_VERSION=1.28.0 \
  GOGO_PROTOBUF_VERSION=1.3.2 \
  GRPC_GATEWAY_VERSION=1.16.0


RUN go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION}
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${GRPC_GATEWAY_VERSION}

# install cosmos gogoproto (includes protoc-gen-gocosmos)
RUN go install github.com/cosmos/gogoproto/protoc-gen-gogo@latest
RUN go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest

COPY --from=BUILDER /usr/local/bin /usr/local/bin
