package main

import (
	"encoding/json"
	"log"
)

type Input struct {
	Address  string  `json:"address"`
	WalletId string  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
}

type Output struct {
	Address  string  `json:"address"`
	WalletId string  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
}

type Tx struct {
	Time     uint64   `json:"time"`
	WalletId string   `json:"wallet_id"`
	Label    string   `json:"label"`
	In       []Input  `json:"in"`
	Out      []Output `json:"out"`
}

func GetTx(query string) (tx Tx) {
	log.Println("http: get tx", query)
	url := ApiUrl + "/tx?txid=" + query + "&caller=" + ApiAgent
	txsIn := []Input{}

	bytes := HttpRequest(url)
	_ = json.Unmarshal(bytes, &tx)

	for _, in := range tx.In {
		in.WalletId = tx.WalletId
		txsIn = append(txsIn, in)
	}
	tx.In = txsIn
	return
}
