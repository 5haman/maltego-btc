package main

import (
	"github.com/albrow/zoom"
)

type Address struct {
	Found					bool			`json:"found"`
	Address       string		`json:"address"`
	Label     		string		`json:"label,omitempty"`
	WalletId			string		`json:"wallet_id"`
	TxCount				int 			`json:"txs_count"`
	TxList			  []ShortTx `json:"txs"`
}

type Wallet struct {
	Found					bool			`json:"found"`
	Label     		string		`json:"label,omitempty"`
	WalletId			string		`json:"wallet_id"`
	TxCount				int 			`json:"txs_count"`
	TxList				[]Tx			`json:"txs"`
}

type Transform struct {
	Type					string
	Direction			string
	Value					string
	LinkColor			string
	Weight				int
	LinkLabel			string
	IconURL				string
	Count					int
}

type ShortTx struct {
	Txid          string	 	`json:"txid"`
	BlockHeight		int			 	`json:"block_height"`
	Time          int64 	 	`json:"time"`
	Sent          float64		`json:"amount_sent"`
	Received      float64		`json:"amount_received"`
	IsInput      	bool			`json:"used_as_input"`
	IsOutput      bool			`json:"used_as_output"`
}

type Tx struct {
	Found					bool			`json:"found"`
	Txid          string	 	`json:"txid"`
	BlockHeight		int			 	`json:"block_height"`
	Time          int64 	 	`json:"time"`
	WalletId			string		`json:"wallet_id"`
	Label     		string	 	`json:"label"`
	Type      		string		`json:"-"`
	In		        []Input 	`json:"in"`
	Out		        []Output 	`json:"out"`
}

type Input struct {
	Address				string		`json:"address"`
	Amount        float64		`json:"amount"`
}

type Output struct {
	Address				string		`json:"address"`
	WalletId			string		`json:"wallet_id"`
	Label     		string		`json:"label"`
	Amount        float64		`json:"amount"`
}

type TransformList struct {
	Address				string
	EntityList		[]Transform
	zoom.Model		`json:"-"`
}

func (list *TransformList) ModelId() string {
	return list.Address
}

func (list *TransformList) SetModelId(id string) {
	list.Address = id
}
