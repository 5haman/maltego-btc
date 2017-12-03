package model

import (
	"log"
	"strconv"
	"encoding/json"

	"github.com/albrow/zoom"
)

type Wallet struct {
	Found					bool			`json:"found"`
	Label     		string		`json:"label"`
	WalletId			string		`json:"wallet_id"`
	TxCount				uint 			`json:"txs_count"`
	Histogram			[]float64	`json:"histogram"`
	AddrTx			  []AddrTx 	`json:"txs"`
	TxList				[]Tx			`json:"tx_list"`
	zoom.Model
}

func (w *Wallet) ModelId() string {
	return w.WalletId
}

func (w *Wallet) SetModelId(id string) {
	w.WalletId = id
}

func GetWallet(query string) (wallet Wallet) {
	err := WalletModel.Find(query, &wallet)
	if err != nil {
		if _, ok := err.(zoom.ModelNotFoundError); ok {
			wallet = RequestWallet(query, 0)

			if wallet.TxCount > step && wallet.TxCount < config.TxsThreshold {
				for from := step; from <= wallet.TxCount; from += step {
					wallet2 := RequestWallet(query, from)
					for _, tx := range wallet2.TxList {
						wallet.TxList = append(wallet.TxList, tx)
					}
				}
			}

			x := TimeRange{}
			for _, t := range wallet.AddrTx {
				tx := GetTx(t.Txid)
				h := tx.Time % 24
				wallet.TxList = append(wallet.TxList, tx)
				x = append(x, float64(h))
			}
			wallet.AddrTx = wallet.AddrTx[:0]

			wallet.Histogram = HourHistogram(x)

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
	return
}

func RequestWallet(query string, from uint) (wallet Wallet) {
	url := config.ApiUrl + "/wallet?wallet=" + query + "&from=" + strconv.Itoa(int(from)) + "&count=100&caller=" + config.ApiAgent

	bytes := HttpRequest(url)
	_ = json.Unmarshal(bytes, &wallet)
	return
}
