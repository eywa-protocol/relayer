.DEFAULT_GOAL := build

.PHONY: update-modules build key


deps:
	go mod tidy
	go mod download

build:
	go build -o bridge  cmd/node.go

key:
	go run key/keygen.go --prefix $(name)

clean:
	rm -f ./bridge keys/*.key ./eth-contracts -r

all: deps build

.PHONY: docker
develop: clean
	@docker-compose up -d $(servicename);
	@docker-compose logs -f $(servicename)

