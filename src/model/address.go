package model

import (
	"log"
	"strconv"
	"encoding/json"

	"github.com/albrow/zoom"
)

type AddrTx struct {
	Txid          string	 	`json:"txid"`
	BlockHeight		uint			`json:"block_height"`
	Time          uint64 	 	`json:"time"`
	Sent          float64		`json:"amount_sent"`
	Received      float64		`json:"amount_received"`
	IsInput      	bool			`json:"used_as_input"`
	IsOutput      bool			`json:"used_as_output"`
}

type Address struct {
	Found					bool			`json:"found"`
	Address       string		`json:"address"`
	Label     		string		`json:"label"`
	WalletId			string		`json:"wallet_id"`
	TxCount				uint 			`json:"txs_count"`
	Histogram			[]float64	`json:"histogram"`
	AddrTx			  []AddrTx 	`json:"txs"`
	TxList				[]Tx			`json:"tx_list"`
	zoom.Model
}

func (a *Address) ModelId() string {
	return a.Address
}

func (a *Address) SetModelId(id string) {
	a.Address = id
}

func GetAddress(query string) (addr Address) {
	err := AddressModel.Find(query, &addr)
	if err != nil {
		if _, ok := err.(zoom.ModelNotFoundError); ok {
			addr = RequestAddress(query, 0)

			if addr.TxCount > step {
				for from := step; from <= addr.TxCount; from += step {
					addr2 := RequestAddress(query, from)
					for _, tx := range addr2.AddrTx  {
						addr.AddrTx  = append(addr.AddrTx , tx)
					}
				}
			}

			x := TimeRange{}
			for _, t := range addr.AddrTx {
				tx := GetTx(t.Txid)
				h := tx.Time % 24
				x = append(x, float64(h))
				addr.TxList = append(addr.TxList, tx)
			}
			addr.AddrTx = addr.AddrTx[:0]

			addr.Histogram = HourHistogram(x)
		} else {
			log.Println("cache error:", err)
		}
	} else {
		log.Println("cache hit:", query)
	}
	return
}

func RequestAddress(query string, from uint) (addr Address) {
	url := config.ApiUrl + "/address?address=" + query + "&from=" + strconv.Itoa(int(from)) + "&count=100&caller=" + config.ApiAgent

	bytes := HttpRequest(url)
	_ = json.Unmarshal(bytes, &addr)
	return
}
