FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

ADD ./eth-contracts/wrappers /eth-contracts/wrappers

WORKDIR /p2p-bridge-b

ADD    ./p2p-bridge .

RUN make

FROM golang:alpine

COPY --from=build /p2p-bridge-b/bridge ./

COPY --from=build /p2p-bridge-b/$TYPE_ADAPTER_ENV ./

RUN apk add --update curl && apk add jq && rm -rf /var/cache/apk/*

EXPOSE $PORT

ENTRYPOINT  export SCALED_NUM=$(curl -s -XGET --unix-socket /var/run/docker.sock -H "Content-Type: application/json" http://v1.24/containers/$(hostname)/json | jq -r .Name[1:]) && \
./bridge -mode init -cnf $TYPE_ADAPTER_ENV && ./bridge -cnf $TYPE_ADAPTER_ENV

