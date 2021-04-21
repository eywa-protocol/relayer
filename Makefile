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
	rm ./bridge keys/*.key

all: deps keys build


boot_key:
	go run key/keygen.go --prefix boot
