FROM golang:alpine as build
ARG NAMEFILE_KEY

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

ADD ./eth-contracts/wrappers /eth-contracts/wrappers

WORKDIR /p2p-bridge-b
ADD    ./p2p-bridge .
#... add conf form ganache

RUN name=${NAMEFILE_KEY} make key
RUN make all

FROM golang:alpine

COPY --from=build /p2p-bridge-b/bridge ./
COPY --from=build /p2p-bridge-b/keys/*.key  keys/
COPY --from=build /p2p-bridge-b/*.env ./


#EXPOSE ${PORT}
ENTRYPOINT ./bridge -cnf ${TYPE_ADAPTER_ENV}
