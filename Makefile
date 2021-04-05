install_dir := $(shell find "$(HOME)/Library/Application Support/maltego" -type d -maxdepth 1 | tail -n 1)

default: build

build:
	mkdir -p ./build
	go build -o ./build/maltego-btc ./
	strip -x ./build/maltego-btc

install:
	@echo "Maltego directory: $(install_dir)"
	@cp -vf ./build/maltego-btc /usr/local/bin/maltego-btc

.PHONY: build
