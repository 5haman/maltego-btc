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
	defFolder string = "maltego-btc"
)

func main() {
	// makes config folder
	appconfig, _ := os.UserConfigDir()
	_ = os.MkdirAll(appconfig+
		string(os.PathSeparator)+
		defFolder, 0755)
	// if path is defined use it, otherwise fall back to default
	LogFile := config.LogFile
	if LogFile == "" {
		LogFile = appconfig +
			string(os.PathSeparator) +
			defFolder +
			string(os.PathSeparator) +
			"maltego-btc.log"
	}
	CacheFile := config.CacheFile
	if CacheFile == "" {
		LogFile = appconfig +
			string(os.PathSeparator) +
			defFolder +
			string(os.PathSeparator) +
			"maltego-btc.cache"
	}
	argc := parseArgs()
	config = ParseConfig(path)

	// enable logging
	f, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	if argc >= 3 {
		LoadCache(CacheFile)
		list := GetTransform(query, Type)

		FilterTransform(query, Type, &list)
		PrintTransform(&list)
		CacheGC()
		SaveCache(CacheFile)
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func parseArgs() (argc int) {
	appconfig, _ := os.UserConfigDir()
	defConfig := appconfig +
		string(os.PathSeparator) +
		defFolder +
		string(os.PathSeparator) +
		"maltego-btc.conf"

	configTemplate := `
{
  "logfile":  "",
  "cachefile":  "",
  "link_address_color": "#B0BEC5",
  "link_wallet_color": "#107896",
  "wallet_max_size": 5000,
  "cache_addresses": 1000,
  "cache_wallets": 5000,
  "icon_address": "https://raw.githubusercontent.com/Megarushing/maltego-btc/master/assets/bitcoin.png",
  "icon_wallet": "https://raw.githubusercontent.com/Megarushing/maltego-btc/master/assets/wallet.png",
  "icon_service": "https://raw.githubusercontent.com/Megarushing/maltego-btc/master/assets/service.png"
}
`
	if _, err := os.Stat(defConfig); err != nil {
		f, _ := os.Create(defConfig)
		f.WriteString(configTemplate)
		f.Close()
	}
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
