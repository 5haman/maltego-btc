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
	ApiAgent			string	`json:"ua"`
	ApiUrl				string	`json:"api_url"`
  LogFile				string	`json:"logfile"`
  RedisURL			string	`json:"redis_url"`
	IconAddress		string	`json:"icon_address"`
  IconWallet		string	`json:"icon_wallet"`
  IconService		string	`json:"icon_service"`
	LinkColor1		string	`json:"link_default"`
	LinkColor2		string	`json:"link_service"`
	TxsThreshold 	uint		`json:"threshold"`
}

type TimeRange []float64

var (
	AddressModel *zoom.Collection
	TxModel	*zoom.Collection
	WalletModel	*zoom.Collection
	TransformModel *zoom.Collection
	pool *zoom.Pool
	step uint = 100
	config Config
)

func ParseConfig(path string) (conf Config){
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("main: Can't open config file '" + path + "'")
		os.Exit(1)
	}
	content := string(file)
	json.Unmarshal([]byte(content), &conf)
	config = conf
	return
}

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

func InitCache() {
	var err error

	pool = zoom.NewPool(config.RedisURL)
	opt := zoom.DefaultCollectionOptions.WithIndex(true)

	AddressModel, err = pool.NewCollectionWithOptions(&Address{}, opt)
	if err != nil {
		log.Println(err)
	}

	TxModel, err = pool.NewCollectionWithOptions(&Tx{}, opt)
	if err != nil {
		log.Println(err)
	}

	TransformModel, err = pool.NewCollectionWithOptions(&TransformList{}, opt)
	if err != nil {
		log.Println(err)
	}

	WalletModel, err = pool.NewCollectionWithOptions(&Wallet{}, opt)
	if err != nil {
		log.Println(err)
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
