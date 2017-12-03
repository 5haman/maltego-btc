package main

import (
	"os"
	"flag"
	"log"
	"fmt"

	"./model"

	"github.com/glennzw/maltegogo"
)

var (
	query 	string
	Type  	string
	path  	string
	help  	bool
	config 	model.Config
	defConfig string = "/usr/local/etc/mbtc.conf"
)

func main() {
	// parse flags and config
	argc := parseArgs()
	config = model.ParseConfig(path)

	// enable logging
	f, err := os.OpenFile(config.LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
			fmt.Println("error opening file: %v", err)
			os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	if argc >= 3 {
		log.Println("args:", os.Args)
		model.InitCache()
		list := model.GetTransform(query, Type)

		model.FilterTransform(query, Type, &list)
		model.PrintTransform(&list)
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

	// get transform type
	Type = os.Args[1]
	os.Args = append(os.Args[:0], os.Args[1:]...)

	// parse input params
	lt := maltegogo.ParseLocalArguments(os.Args)
	query = lt.Value

	return
}
