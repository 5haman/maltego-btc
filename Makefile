
default: build install


build:
	mkdir -p ./build
	go build -o ./build/mbtc ./src/

install:
	cp -f ./build/mbtc /usr/local/bin/mbtc

.PHONY: build
