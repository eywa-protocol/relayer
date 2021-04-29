FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

ADD ./eth-contracts/wrappers /eth-contracts/wrappers

WORKDIR /p2p-bridge-b

ADD    ./p2p-bridge .

RUN make  deps build

FROM golang:alpine

WORKDIR /p2p-bridge

COPY --from=build /p2p-bridge-b/bridge ./

COPY --from=build /p2p-bridge-b/${TYPE_ADAPTER_ENV} ./

RUN echo "CONFIG ${TYPE_ADAPTER_ENV}"

RUN ./bridge -mode init -cnf ${TYPE_ADAPTER_ENV}
#EXPOSE ${PORT}
ENTRYPOINT ./bridge -cnf ${TYPE_ADAPTER_ENV}
