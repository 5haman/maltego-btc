package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/glennzw/maltegogo"
)

var (
	query     string
	Type      string
	path      string
	help      bool
	defConfig string = "/usr/local/etc/mbtc.conf"
)

func main() {
	// parse flags and config
	argc := parseArgs()
	config = ParseConfig(path)

	// enable logging
	f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	if argc >= 3 {
		LoadCache(config.CacheFile)
		list := GetTransform(query, Type)

		FilterTransform(query, Type, &list)
		PrintTransform(&list)
		CacheGC()
		SaveCache(config.CacheFile)
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func parseArgs() (argc int) {
	// parse flags
	flag.StringVar(&path, "c", defConfig, "Path to config file")
	flag.BoolVar(&help, "h", false, "Print Help")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	argc = len(os.Args)
	//fmt.Println("Args:", os.Args)

	// get transform type
	Type = os.Args[1]
	os.Args = append(os.Args[:0], os.Args[1:]...)

	// parse input params
	lt := maltegogo.ParseLocalArguments(os.Args)
	query = lt.Value

	return
}
