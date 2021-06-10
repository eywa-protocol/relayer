.DEFAULT_GOAL := build

.PHONY: update-modules build key

deps:
	go mod tidy
	go mod download

build:	deps
	go build -o bridge  cmd/main.go

clean:
	rm -f ./bridge keys/*.key keys/*.env

all: deps build

.PHONY: docker
develop:
	@docker-compose up -d $(servicename);
	@docker-compose logs -f $(servicename)

bls_test:
	 go test -v ./libp2p/pub_sub_bls/libp2p_pubsub -run TestBLS

custom_bls_test:
	 go test -v ./libp2p/pub_sub_bls/libp2p_pubsub -run TestOneStepBLS

