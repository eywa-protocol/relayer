FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

ADD ./external/eth-contracts/wrappers /eth-contracts/wrappers

WORKDIR /p2p-bridge-b

ADD    ./p2p-bridge .

RUN make

FROM golang:alpine

COPY --from=build /p2p-bridge-b/bridge ./

COPY --from=build /p2p-bridge-b/$TYPE_ADAPTER_ENV ./

EXPOSE $PORT

ENTRYPOINT ./bridge -mode init -cnf $TYPE_ADAPTER_ENV && sleep 5 && ./bridge -randevoue $RANDEVOUE -cnf $TYPE_ADAPTER_ENV

