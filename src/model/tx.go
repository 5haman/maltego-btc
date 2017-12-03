package model

import (
	"log"
	"encoding/json"

	"github.com/albrow/zoom"
)

type Input struct {
	Address				string		`json:"address"`
	WalletId			string		`json:"wallet_id"`
	Label     		string		`json:"label"`
	Amount        float64		`json:"amount"`
}

type Output struct {
	Address				string		`json:"address"`
	WalletId			string		`json:"wallet_id"`
	Label     		string		`json:"label"`
	Amount        float64		`json:"amount"`
}

type Tx struct {
	Found					bool			`json:"found"`
	Txid          string	 	`json:"txid"`
	BlockHeight		uint			`json:"block_height"`
	Time          uint64 	 	`json:"time"`
	WalletId			string		`json:"wallet_id"`
	Label     		string	 	`json:"label"`
	Type      		string		`json:"type"`
	In		        []Input 	`json:"in"`
	Out		        []Output 	`json:"out"`
	zoom.Model
}

func (tx *Tx) ModelId() string {
	return tx.Txid
}

func (tx *Tx) SetModelId(id string) {
	tx.Txid = id
}

func GetTx(query string) (tx Tx) {
	err := TxModel.Find(query, &tx)
	if err != nil {
		if _, ok := err.(zoom.ModelNotFoundError); ok {
			RequestTx(query, &tx)

			// save to redis cache
			if err := TxModel.Save(&tx); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("cache error:", err)
		}
	}
	
	return
}

func RequestTx(query string, tx *Tx) {
  url := config.ApiUrl + "/tx?txid=" + query + "&caller=" + config.ApiAgent
	txsIn := []Input{}

	bytes := HttpRequest(url)
	_ = json.Unmarshal(bytes, &tx)

	for _, in := range tx.In {
		in.WalletId = tx.WalletId
		if (tx.Label != "") {
			in.Label = tx.Label
		}
		txsIn = append(txsIn, in)
	}
	tx.In = txsIn
	return
}
