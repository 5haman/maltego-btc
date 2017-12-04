package model

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/albrow/zoom"
)

type AddrTx struct {
	Txid string `json:"txid"`
	Time uint64 `json:"time"`
}

type Address struct {
	Address   string    `json:"address"`
	Label     string    `json:"label"`
	WalletId  string    `json:"wallet_id"`
	TxCount   int       `json:"txs_count"`
	Histogram []float64 `json:"histogram"`
	AddrTx    []AddrTx  `json:"txs" redis:"-"`
	TxList    []Tx      `json:"tx_list"`
	Cached    uint64    `json:"-" zoom:"index"`
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
					for _, tx := range addr2.AddrTx {
						addr.AddrTx = append(addr.AddrTx, tx)
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
			addr.Cached = uint64(time.Now().Unix())

			// save to redis cache
			if err := AddressModel.Save(&addr); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("cache error:", err)
		}
	}
	return
}

func RequestAddress(query string, from int) (addr Address) {
	// check if this request already fired
	if requestMap[query] == true {
		// wait and get it from cache
		go func() {
			for {
				if err := AddressModel.Find(query, &addr); err == nil {
					return
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	} else {
		requestMap[query] = true
		log.Println("http: get", query)

		url := ApiUrl + "/address?address=" + query + "&from=" + strconv.Itoa(int(from)) + "&count=100&caller=" + ApiAgent

		bytes := HttpRequest(url)
		_ = json.Unmarshal(bytes, &addr)
	}
	return
}
