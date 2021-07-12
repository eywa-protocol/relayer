.DEFAULT_GOAL := build-all

.PHONY: update-modules build key test_local_bridge

deps:
	go mod tidy
	go mod download

build:	deps
	go build -o bridge  cmd/bridge/bridge.go

.PHONY: build-bsn

build-bsn: deps
	go build -o bsn  cmd/bsn/bsn.go

clean:
	rm -f ./bridge ./bsn keys/*.key keys/*.env

build-all: deps build build-bsn

.PHONY: docker
develop:
	@docker-compose up -d $(servicename);
	@docker-compose logs -f $(servicename)

bls_test:
	 go test -v ./libp2p/pub_sub_bls/libp2p_pubsub -run TestBLS

custom_bls_test:
	 go test -v ./libp2p/pub_sub_bls/libp2p_pubsub -run TestOneStepBLS

test_local_bridge:
	go test -v ./test -run Test_SendRequestV2
