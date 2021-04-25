.DEFAULT_GOAL := build

.PHONY: update-modules build key


deps:
	go mod tidy
	go mod download

build: keys
	go build -o bridge  cmd/node.go

key:
	go run key/keygen.go --prefix $(name)

clean:
	rm -f ./bridge keys/*.key

all: deps keys build

.PHONY: docker
develop:
	go run key/keygen.go --prefix $(keyname);
	@docker-compose up -d $(servicename);
	@docker-compose logs -f $(servicename)

