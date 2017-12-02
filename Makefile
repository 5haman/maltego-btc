
default: build install


build:
	mkdir -p ./build
	go build -o ./build/mbtc ./src
	strip -x ./build/mbtc

install:
	cp -f ./build/mbtc /usr/local/bin

.PHONY: build
