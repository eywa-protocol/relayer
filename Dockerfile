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
COPY    ./forward ./forward
COPY    ./helpers ./helpers
COPY    ./libp2p ./libp2p
COPY    ./node ./node
COPY    ./prom ./prom
COPY    ./runa ./runa
COPY    ./Makefile .
COPY    ./sentry ./sentry
COPY    ./.git ./.git

RUN APP_VERSION=$(git tag --sort=taggerdate | tail -1) && \
    COMMIT_SHA=$(git rev-list -1 HEAD) && \
    BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
    go install -ldflags="-X gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common.Version=$APP_VERSION -X gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common.Commit=$COMMIT_SHA -X gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common.BuildTime=$BUILD_TIME"  ./cmd/...

FROM golang:alpine

#create mounting points for volumes
RUN mkdir -p ./keys
RUN touch ./bsn.yaml
RUN touch ./bootstrap.env

COPY --from=build /go/bin/bridge ./

COPY --from=build /go/bin/bsn ./

COPY --from=build /go/bin/gsn ./
