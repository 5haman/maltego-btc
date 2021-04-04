install_dir := $(shell find "$(HOME)/Library/Application Support/maltego" -type d -maxdepth 1 | tail -n 1)

default: build

build:
	mkdir -p ./build
	go build -o ./build/mbtc ./
	strip -x ./build/mbtc

install:
	@echo "Maltego directory: $(install_dir)"
	@cp -vf ./build/mbtc /usr/local/bin/mbtc
	@cp -vf ./config.json /usr/local/etc/mbtc.conf


.PHONY: build
