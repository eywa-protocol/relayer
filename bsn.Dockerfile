FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

WORKDIR /p2p-bridge-b

COPY    ./p2p-bridge/go.mod .
COPY    ./p2p-bridge/go.sum .
ADD     ./p2p-bridge/external/eth-contracts/wrappers ./external/eth-contracts/wrappers

RUN go mod download

COPY    ./p2p-bridge/cmd ./cmd
COPY    ./p2p-bridge/common ./common
COPY    ./p2p-bridge/config ./config
COPY    ./p2p-bridge/helpers ./helpers
COPY    ./p2p-bridge/libp2p ./libp2p
COPY    ./p2p-bridge/Makefile .
COPY    ./p2p-bridge/bootstrap.env .
COPY    ./p2p-bridge/node ./node

RUN mkdir -p ./keys

RUN make

FROM golang:alpine

COPY --from=build /p2p-bridge-b/bridge ./

COPY --from=build /p2p-bridge-b/$TYPE_ADAPTER_ENV ./

EXPOSE $PORT

ENTRYPOINT ./bridge -mode bs-init -cnf $TYPE_ADAPTER_ENV && sleep 5 && ./bridge -mode bs-start -cnf $TYPE_ADAPTER_ENV

