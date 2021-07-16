FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

WORKDIR /p2p-bridge-b
COPY    ./go.mod .
COPY    ./go.sum .
ADD     ./external/eth-contracts/wrappers ./external/eth-contracts/wrappers

RUN go mod download

COPY    ./cmd ./cmd
COPY    ./common ./common
COPY    ./config ./config
COPY    ./helpers ./helpers
COPY    ./libp2p ./libp2p
COPY    ./node ./node
COPY    ./runa ./runa
COPY    ./Makefile .

RUN make

FROM golang:alpine

#create mounting points for volumes
RUN mkdir -p ./keys
RUN touch ./bsn.yaml
RUN touch ./bootstrap.env

COPY --from=build /p2p-bridge-b/bridge ./

COPY --from=build /p2p-bridge-b/bsn ./
