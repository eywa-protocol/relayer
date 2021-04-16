.DEFAULT_GOAL := build

.PHONY: update-modules build keys


deps: update-modules
	go mod tidy
	go mod download
link:
	ln -s /home/syi/src/simplifi/eth-contracts/wrappers/ external/

build: keys
	go build -o bridge  cmd/node.go

keys:
	go run key/keygen.go --prefix srv3

clean:
	rm ./bridge keys/*.key

all: deps keys build

update-modules:
	git submodule update --init --recursive
	touch $@

clean-modules:
	rm ./external/wrappers/*.go
