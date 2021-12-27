package main

import (
	"encoding/gob"
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
	CacheAddresses   int    `json:"cache_addresses"`
	CacheWallets     int    `json:"cache_wallets"`
}

type TimeRange []float64

const WalletURL = "https://www.walletexplorer.com/wallet/"
const ApiUrl = "https://www.walletexplorer.com/api/1"
const ApiAgent = "maltego-btc"
const step = 100

var (
	WalletModel          *cache.Cache
	WalletAddressesModel *cache.Cache
	config               Config
	requestMap           = map[string]bool{}
)

type CachedInterface interface {
	GetId() string
	GetCacheTime() uint64
}

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

func LoadCache(fname string) error {
	gob.Register(WalletAddresses{})
	gob.Register(Wallet{})
	fp, err := os.Open(fname)
	if err != nil {
		WalletAddressesModel = cache.New(24*time.Hour, 48*time.Hour)
		WalletModel = cache.New(24*time.Hour, 48*time.Hour)
		log.Println("no cache found, making new...", err)
		return err
	}
	dec := gob.NewDecoder(fp)
	caches := []map[string]cache.Item{}
	err = dec.Decode(&caches)
	if err == nil {
		WalletModel = cache.NewFrom(24*time.Hour, 48*time.Hour, caches[0])
		WalletAddressesModel = cache.NewFrom(24*time.Hour, 48*time.Hour, caches[1])
	}
	if err != nil {
		fp.Close()
		WalletAddressesModel = cache.New(24*time.Hour, 48*time.Hour)
		WalletModel = cache.New(24*time.Hour, 48*time.Hour)
		log.Println("error loading cache, making new...", err)
		return err
	}
	return fp.Close()
}

func SaveCache(fname string) error {
	caches := []map[string]cache.Item{
		WalletModel.Items(),
		WalletAddressesModel.Items(),
	}
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(fp)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	for _, c := range caches {
		gob.Register(c)
		for _, i := range c {
			gob.Register(i.Object)
		}
	}
	err = enc.Encode(&caches)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func CleanupModel(Model *cache.Cache, amount int) {
	q := Model.Items()
	removed := 0

	values := make([]CachedInterface, 0, len(q))
	for _, v := range q {
		values = append(values, v.Object.(CachedInterface))
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].GetCacheTime() < values[j].GetCacheTime()
	})

	for _, obj := range values {
		Model.Delete(obj.GetId())
		removed++
		if removed > amount {
			break
		}
	}

	if removed > 0 {
		log.Println("gc:", removed, "old objects deleted")
	}
}

func CacheGC() {
	addrCount := len(WalletAddressesModel.Items())
	walletCount := len(WalletModel.Items())

	// delete old addresses
	addrToRemove := addrCount - config.CacheAddresses
	if addrToRemove > 0 {
		CleanupModel(WalletAddressesModel, addrToRemove)
	}
	// delete old wallets
	walletsToRemove := walletCount - config.CacheWallets
	if walletsToRemove > 0 {
		CleanupModel(WalletModel, walletsToRemove)
	}
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
