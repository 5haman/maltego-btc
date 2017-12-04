package model

import (
	"fmt"
	"log"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"sort"
	"os"

	"github.com/albrow/zoom"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

type Config struct {
	ApiAgent				string	`json:"ua"`
	ApiUrl					string	`json:"api_url"`
  LogFile					string	`json:"logfile"`
  RedisURL				string	`json:"redis_url"`
	IconAddress			string	`json:"icon_address"`
  IconWallet			string	`json:"icon_wallet"`
  IconService			string	`json:"icon_service"`
	LinkColor1			string	`json:"link_default"`
	LinkColor2			string	`json:"link_service"`
	TxsThreshold 		int			`json:"wallet_max_size"`
	CacheAddresses	uint		`json:"cache_addresses"`
	CacheWallets		uint		`json:"cache_wallets"`
}

type TimeRange []float64

const ApiUrl = "https://www.walletexplorer.com/api/1"
const ApiAgent = "maltego-btc"
const step = 100

var (
	AddressModel *zoom.Collection
	WalletModel	*zoom.Collection
	pool *zoom.Pool
	config Config
)

func ParseConfig(path string) (conf Config){
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
	var err error

	pool = zoom.NewPool(config.RedisURL)
	opt := zoom.DefaultCollectionOptions.WithIndex(true).WithName("a")

	AddressModel, err = pool.NewCollectionWithOptions(&Address{}, opt)
	if err != nil {
		log.Println(err)
	}

	opt = zoom.DefaultCollectionOptions.WithIndex(true).WithName("w")
	WalletModel, err = pool.NewCollectionWithOptions(&Wallet{}, opt)
	if err != nil {
		log.Println(err)
	}
}

func CacheGC() {
	addrCount := 0
	walletCount := 0
	removed := 0
	a := []*Address{}
	w := []*Wallet{}

	t := pool.NewTransaction()
	t.Count(AddressModel, &addrCount)
	t.Count(WalletModel, &walletCount)

	if err := t.Exec(); err != nil {
		log.Println(err)
	}

	// delete old addresses
	addrToRemove := uint(addrCount) - config.CacheAddresses
	if addrToRemove > 0 {
		q := AddressModel.NewQuery().Include("Address").Order("-Cached").Offset(config.CacheAddresses)
		if err := q.Run(&a); err != nil {
			log.Println(err)
		}
		for id := range a {
			if _, err := AddressModel.Delete(a[id].Address); err != nil {
				log.Println(err)
			}
			removed++
		}
	}

	// delete old wallets
	walletsToRemove := uint(walletCount) - config.CacheWallets
	if walletsToRemove > 0 {
		q := WalletModel.NewQuery().Include("WalletId").Order("-Cached").Offset(config.CacheWallets)
		if err := q.Run(&w); err != nil {
			log.Println(err)
		}
		for id := range w {
			log.Println(w[id])
			if _, err := WalletModel.Delete(w[id].WalletId); err != nil {
				log.Println(err)
			}
			removed++
		}
	}

	if removed > 0 {
		log.Println("gc:", removed, "old objects deleted")
	}
}

func ClosePool() error {
	return pool.Close()
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
