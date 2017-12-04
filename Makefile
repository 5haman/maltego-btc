install_dir := $(shell find "$(HOME)/Library/Application Support/maltego" -type d -maxdepth 1 | tail -n 1)

default: deps build

deps:
	go get "github.com/glennzw/maltegogo"
	go get "github.com/albrow/zoom"
	go get "gonum.org/v1/gonum/floats"
	go get "gonum.org/v1/gonum/stat"

build:
	mkdir -p ./build
	go build -o ./build/mbtc ./src
	strip -x ./build/mbtc

install:
	@echo "Maltego directory: $(install_dir)"
	@cp -Rv maltego/* "$(install_dir)/config/Maltego/"
	@cp -vf ./build/mbtc /usr/local/bin/mbtc
	@cp -vf ./config.json /usr/local/etc/mbtc.conf


.PHONY: build
