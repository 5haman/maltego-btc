package main

import (
	"os"
	"flag"
	"log"
	"fmt"

	mt "github.com/glennzw/maltegogo"
)

var (
	path string
	help 	bool
)

func main() {
	// parse flags
	flag.StringVar(&path, "c", "", "Path to config file")
	flag.BoolVar(&help, "h", false, "Print Help")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	//readConfig(path)

	argc := len(os.Args)
	if argc >= 3 {
		// enable logging
		f, err := os.OpenFile("/usr/local/var/log/maltego-btc.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
		if err != nil {
		    fmt.Println("error opening file: %v", err)
				os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)

		log.Println("os.args:", os.Args)

		// get transform type
		Type := os.Args[1]
	 	os.Args= append(os.Args[:0], os.Args[1:]...)

		// parse input params
	  lt := mt.ParseLocalArguments(os.Args)
		input := lt.Value

		// run transform
		InitCache()
		tr := RunTransform(input, Type)
		PrintTransform(&tr)
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
