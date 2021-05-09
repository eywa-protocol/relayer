.DEFAULT_GOAL := build

.PHONY: update-modules build key


deps:
	go mod tidy
	go mod download

build:
	go build -o bridge  cmd/main.go

clean:
	rm -f ./bridge keys/*.key

all: deps build

.PHONY: docker
develop:
	@docker-compose up -d $(servicename);
	@docker-compose logs -f $(servicename)

