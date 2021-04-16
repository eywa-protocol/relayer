FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

WORKDIR /p2p-bridge

ADD . .

ADD ./env_p2p_bridge.env .

RUN make

FROM golang:alpine

COPY --from=build /p2p-bridge/bridge /bridge

COPY --from=build /p2p-bridge/keys/srv3-ecdsa.key  keys/

COPY --from=build /p2p-bridge/env_p2p_bridge.env .

EXPOSE ${PORT}

ENTRYPOINT ["/bridge"]
