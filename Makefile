.DEFAULT_GOAL := build-all

.PHONY: update-modules build key test_local_bridge build-bsn

deps:
	go mod tidy
	go mod download

build-bsn: deps
	go build -o bsn  cmd/bsn/bsn.go
build: deps
	go build -o bridge  cmd/bridge/bridge.go


clean:
	rm -f ./bridge ./bsn;rm -rf ./.data/*

build-all: deps build build-bsn

.PHONY: docker
develop:
	@docker-compose up -d $(servicename);
	@docker-compose logs -f $(servicename)

bls_test:
	 go test -v ./test/ -run TestBLS

custom_bls_test:
	 go test -v ./test/ -run TestOneStepBLS

test_local_bridge:
	go test -v ./test/ -run Test_Local_SendRequestV2 $(date +%s)

wrappers:
	make -C external/eth-contracts/ eth-local-migrate

gen_proto:
	protoc --proto_path=consensus/protobuf/message --go_out=plugins=grpc:consensus/protobuf/message --go_opt=paths=source_relative consensus/protobuf/message/message.proto
