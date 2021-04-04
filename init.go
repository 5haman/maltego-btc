package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/patrickmn/go-cache"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

type Config struct {
	ApiAgent         string `json:"ua"`
	ApiUrl           string `json:"api_url"`
	LogFile          string `json:"logfile"`
	CacheFile        string `json:"cachefile"`
	IconAddress      string `json:"icon_address"`
	IconWallet       string `json:"icon_wallet"`
	IconService      string `json:"icon_service"`
	LinkAddressColor string `json:"link_address_color"`
	LinkWalletColor  string `json:"link_wallet_color"`
	TxsThreshold     int    `json:"wallet_max_size"`
	CacheAddresses   uint   `json:"cache_addresses"`
	CacheWallets     uint   `json:"cache_wallets"`
}

type TimeRange []float64

const ApiUrl = "https://www.walletexplorer.com/api/1"
const ApiAgent = "maltego-btc"
const step = 100

var (
	WalletModel          *cache.Cache
	WalletAddressesModel *cache.Cache
	config               Config
	requestMap           = map[string]bool{}
)

func ParseConfig(path string) (conf Config) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error: %v", err)
		os.Exit(1)
	}
	content := string(file)
	json.Unmarshal([]byte(content), &conf)
	config = conf
	return
}

func InitCache() {
	WalletAddressesModel = cache.New(5*time.Minute, 10*time.Minute)
	WalletModel = cache.New(5*time.Minute, 10*time.Minute)
}

func CacheGC() {
	//addrCount := 0
	//walletCount := 0
	//removed := 0
	//a := []*WalletAddresses{}
	//w := []*Wallet{}
	//
	//t := pool.NewTransaction()
	//t.Count(WalletAddressesModel, &addrCount)
	//t.Count(WalletModel, &walletCount)
	//
	//if err := t.Exec(); err != nil {
	//	log.Println(err)
	//}
	//
	//// delete old addresses
	//addrToRemove := uint(addrCount) - config.CacheAddresses
	//if addrToRemove > 0 {
	//	q := WalletAddressesModel.NewQuery().Include("WalletId").Order("-Cached").Offset(config.CacheAddresses)
	//	if err := q.Run(&a); err != nil {
	//		log.Println(err)
	//	}
	//	for id := range a {
	//		if _, err := WalletAddressesModel.Delete(a[id].WalletId); err != nil {
	//			log.Println(err)
	//		}
	//		removed++
	//	}
	//}
	//
	//// delete old wallets
	//walletsToRemove := uint(walletCount) - config.CacheWallets
	//if walletsToRemove > 0 {
	//	q := WalletModel.NewQuery().Include("WalletId").Order("-Cached").Offset(config.CacheWallets)
	//	if err := q.Run(&w); err != nil {
	//		log.Println(err)
	//	}
	//	for id := range w {
	//		log.Println(w[id])
	//		if _, err := WalletModel.Delete(w[id].WalletId); err != nil {
	//			log.Println(err)
	//		}
	//		removed++
	//	}
	//}
	//
	//if removed > 0 {
	//	log.Println("gc:", removed, "old objects deleted")
	//}
}

func HourHistogram(x TimeRange) (hst []float64) {
	sort.Sort(x)

	// Trim the dividers to create 24 buckets
	div := make([]float64, 24)
	floats.Span(div, 0, 24)
	hst = stat.Histogram(nil, div, x, nil)

	return
}

// sort helper functions
func (tr TimeRange) Len() int           { return len(tr) }
func (tr TimeRange) Swap(i, j int)      { tr[i], tr[j] = tr[j], tr[i] }
func (tr TimeRange) Less(i, j int) bool { return tr[i] < tr[j] }

func HttpRequest(url string) (bytes []byte) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("http error:", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("http error:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("http error:", resp.Status)
		return
	}

	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("http error:", err)
	}
	return
}
