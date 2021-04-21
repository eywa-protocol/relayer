FROM golang:alpine as build

RUN apk add --no-cache git gcc musl-dev linux-headers build-base

WORKDIR /p2p-bridge

ADD . .

ADD ./env_p2p_.env .

#COPY /home/syi/src/DigiU/eth-contracts/wrappers ./wrappers

RUN make

FROM golang:alpine

COPY --from=build /p2p-bridge/bridge /bridge

RUN if [[ -z "$BOOTSTRAP" ]] ; then echo BOOTSTRAP Argument not provided ; else echo Argument is $BOOTSTRAP ; fi

COPY --from=build /p2p-bridge/keys/srv3-ecdsa.key  keys/

COPY --from=build /p2p-bridge/env_p2p_.env .

RUN ls -la

EXPOSE ${PORT}

ENTRYPOINT ["/bridge"]
