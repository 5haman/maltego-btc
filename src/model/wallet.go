package model

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/albrow/zoom"
)

type WalletTxOutput struct {
	WalletId string  `json:"wallet_id"`
	Label    string  `json:"label"`
	Amount   float64 `json:"amount"`
}

type WalletTx struct {
	TxId            string           `json:"txid"`
	WalletId        string           `json:"wallet_id"`
	Label           string           `json:"label"`
	Time            uint64           `json:"time"`
	Amount          float64          `json:"amount"`
	Balance         float64          `json:"balance"`
	Type            string           `json:"type"`
	Fee             float64          `json:"fee"`
	WalletTxOutputs []WalletTxOutput `json:"outputs"`
}

type Wallet struct {
	Label     string     `json:"label"`
	WalletId  string     `json:"wallet_id"`
	TxCount   int        `json:"txs_count"`
	WalletTxs []WalletTx `json:"txs"`
	Histogram []float64  `json:"histogram"`
	Addresses []string   `json:"addresses"`
	Cached    uint64     `json:"-" zoom:"index"`
	zoom.Model
}

type WalletAddress struct {
	Address     string  `json:"address"`
	Balance     float64 `json:"balance"`
	IncomingTxs int     `json:"incoming_txs"`
}

type WalletAddresses struct {
	Label          string          `json:"label"`
	WalletId       string          `json:"wallet_id"`
	AddressesCount int             `json:"addresses_count"`
	Addresses      []WalletAddress `json:"addresses"`
	Cached         uint64          `json:"-" zoom:"index"`
	zoom.Model
}

func (w *WalletAddresses) ModelID() string {
	return w.WalletId
}

func (w *WalletAddresses) SetModelID(id string) {
	w.WalletId = id
}

func (w *Wallet) ModelID() string {
	return w.WalletId
}

func (w *Wallet) SetModelID(id string) {
	w.WalletId = id
}

func GetWallet(query string) (wallet Wallet) {
	err := WalletModel.Find(query, &wallet)
	if err != nil {
		if _, ok := err.(zoom.ModelNotFoundError); ok {
			wallet = RequestWallet(query, 0)

			if wallet.TxCount > step && wallet.TxCount < config.TxsThreshold {
				for from := step; from <= wallet.TxCount; from += step {
					time.Sleep(2000 * time.Millisecond)
					wallet2 := RequestWallet(query, from)
					for _, tx := range wallet2.WalletTxs {
						wallet.WalletTxs = append(wallet.WalletTxs, tx)
					}
				}
			}

			x := TimeRange{}
			for _, t := range wallet.WalletTxs {
				h := t.Time % 24
				x = append(x, float64(h))
			}

			wallet.Histogram = HourHistogram(x)
			wallet.Cached = uint64(time.Now().Unix())

			// save to redis cache
			if err := WalletModel.Save(&wallet); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("cache error:", err)
		}
	} else {
		log.Println("cache hit:", query)
	}

	//log.Println("time histogram:", wallet.WalletId, wallet.Histogram)
	return
}

func RequestWallet(query string, from int) (wallet Wallet) {
	// check if this request already fired
	if requestMap[query] == true {
		// wait and get it from cache
		go func() {
			for {
				if err := WalletModel.Find(query, &wallet); err == nil {
					return
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	} else {
		requestMap[query] = true
		log.Println("http: get wallet ", query)

		url := ApiUrl + "/wallet?wallet=" + query + "&from=" + strconv.Itoa(int(from)) + "&count=100&caller=" + ApiAgent

		bytes := HttpRequest(url)
		_ = json.Unmarshal(bytes, &wallet)
	}
	return
}

func GetWalletAddresses(query string) (addresses WalletAddresses) {
	err := WalletAddressesModel.Find(query, &addresses)

	if err != nil {
		if _, ok := err.(zoom.ModelNotFoundError); ok {
			addresses = RequestWalletAddresses(query, 0)

			if addresses.AddressesCount > step && addresses.AddressesCount < config.TxsThreshold {
				for from := step; from <= addresses.AddressesCount; from += step {
					addresses2 := RequestWalletAddresses(query, from)
					for _, addr := range addresses2.Addresses {
						addresses.Addresses = append(addresses.Addresses, addr)
					}
				}
			}

			addresses.Cached = uint64(time.Now().Unix())

			// save to redis cache
			if err := WalletAddressesModel.Save(&addresses); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("cache error:", err)
		}
	} else {
		log.Println("cache hit:", query)
	}
	return
}

func RequestWalletAddresses(query string, from int) (addresses WalletAddresses) {
	// check if this request already fired
	if requestMap[query] == true {
		// wait and get it from cache
		go func() {
			for {
				if err := WalletAddressesModel.Find(query, &addresses); err == nil {
					return
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	} else {
		requestMap[query] = true
		log.Println("http: get wallet addresses ", query)

		url := ApiUrl + "/wallet-addresses?wallet=" + query + "&from=" + strconv.Itoa(int(from)) + "&count=100&caller=" + ApiAgent

		bytes := HttpRequest(url)
		_ = json.Unmarshal(bytes, &addresses)
	}
	return
}

func Address2Wallet(query string, from int) (wallet Wallet) {
	log.Println("http: get address from wallet ", query)

	url := ApiUrl + "/address-lookup?address=" + query + "&from=" + strconv.Itoa(int(from)) + "&count=100&caller=" + ApiAgent

	bytes := HttpRequest(url)
	_ = json.Unmarshal(bytes, &wallet)
	return
}
