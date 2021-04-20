FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

ADD ./eth-contracts/wrappers /eth-contracts/wrappers

WORKDIR /p2p-bridge-b
RUN env
ADD ./p2p-bridge .

ADD ./p2p-bridge/env_p2p_bridge.env .

RUN make all

FROM golang:alpine

COPY --from=build /p2p-bridge-b/bridge /bridge

COPY --from=build /p2p-bridge-b/keys/srv3-ecdsa.key  keys/

COPY --from=build /p2p-bridge-b/env_p2p_bridge.env .

#EXPOSE ${PORT}

ENTRYPOINT ["/bridge"]
