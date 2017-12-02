package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/glennzw/maltegogo"
	"github.com/albrow/zoom"
)

var (
	RedisURL = "127.0.0.1:6379"
	IconURLAddr = "https://5haman.github.io/images/maltego/bitcoin.png"
	IconURLWt = "https://5haman.github.io/images/maltego/wallet.png"
	LinkColor = "#B0BEC5"
	TransformModel	*zoom.Collection
	pool *zoom.Pool
)

func RunTransform(query string, Type string) (list TransformList) {
	// try to get from redis cache
	err := FromCache(query, &list)

	if _, ok := err.(zoom.ModelNotFoundError); ok {
		// no cached version found, request from external API
		list.Id = query
		log.Println("request:", query)

		switch Type {
		case "WalletFull", "WalletInOut", "WalletIn", "WalletOut", "WalletAddr":
			WalletTransform(query, &list)
		case "AddrFull", "AddrInOut", "AddrIn", "AddrOut", "AddrWallet":
			AddressTransform(query, &list)
		default:
			log.Println("error:", "unknown transform type: " + Type)
			return
		}

		log.Println("finish request:", query)

		// save to redis cache
		if err := TransformModel.Save(&list); err != nil {
			fmt.Println(err)
		}
	}

	FilterTransform(query, Type, &list)

	return
}

func FilterTransform(query string, Type string, list *TransformList) {
	list2 := []Transform{}

	for _, ent := range list.EntityList {
		switch Type {
		case "WalletFull", "AddrFull":
			list2 = append(list2, ent)
		case "WalletInOut":
			if ent.Type == "btc.BtcWallet" {
				list2 = append(list2, ent)
			}
		case "WalletIn":
			if ent.Type == "btc.BtcWallet" && ent.Direction == "in" {
				list2 = append(list2, ent)
			}
		case "WalletOut":
			if ent.Type == "btc.BtcWallet" && ent.Direction == "out" {
				list2 = append(list2, ent)
			}
		case "WalletAddr":
			if ent.Type == "btc.BtcAddress" {
				list2 = append(list2, ent)
			}
		case "AddrInOut":
			if ent.Type == "btc.BtcAddress" {
				list2 = append(list2, ent)
			}
		case "AddrIn":
			if ent.Type == "btc.BtcAddress" && ent.Direction == "in" {
				list2 = append(list2, ent)
			}
		case "AddrOut":
			if ent.Type == "btc.BtcAddress" && ent.Direction == "out" {
				list2 = append(list2, ent)
			}
		case "AddrWallet":
			if ent.Type == "btc.BtcWallet" {
				list2 = append(list2, ent)
			}
		}
	}

	list.EntityList = list2
	return
}

func WalletTransform(query string, list *TransformList) {
	c := map[string]int{}
	m := map[string]Transform{}

	wallet := RequestWallet(query, 0)

	if wallet.TxCount > 100 {
		for from := 100; from <= wallet.TxCount; from += 100 {
			wallet2 := RequestWallet(query, from)

			for _, tx := range wallet2.TxList {
				wallet.TxList = append(wallet.TxList, tx)
			}
		}
	}

	for _, t := range wallet.TxList {
		tx := RequestTx(t.Txid)
		if tx.WalletId == query {
			// Add links to wallet addresses
			for _, in := range tx.In {
				if c[in.Address] == 0 {
					m[in.Address] = Transform{"btc.BtcAddress", "out", in.Address, LinkColor, 100, strconv.FormatFloat(in.Amount, 'f', -1, 64) + " BTC", IconURLAddr, 1}
				}
				c[in.Address]++
			}

			// Add links to other wallets
			for _, out := range tx.Out {
				if out.WalletId != query && c[out.WalletId] == 0 {
					m[out.WalletId] = Transform{"btc.BtcWallet", "out", out.WalletId, LinkColor, 100, strconv.FormatFloat(out.Amount, 'f', -1, 64) + " BTC", IconURLWt, 1}
				}
				c[out.WalletId]++
			}
		} else {
			// Add incoming links to other wallets
			if c[tx.WalletId] == 0 {
				m[tx.WalletId] = Transform{"btc.BtcWallet", "in", tx.WalletId, LinkColor, 100, "", IconURLWt, 1}
			}
			c[tx.WalletId]++
		}
	}

	for _, NewEnt := range m {
		list.EntityList = append(list.EntityList, NewEnt)
	}
}

func AddressTransform(query string, list *TransformList) {
	c := map[string]int{}
	m := map[string]Transform{}

	addr := RequestAddress(query, 0)

	if addr.TxCount > 100 {
	  for from := 100; from <= addr.TxCount; from += 100 {
			addr2 := RequestAddress(query, from)

			for _, tx := range addr2.TxList {
				addr.TxList = append(addr.TxList, tx)
			}
		}
	}

	for _, t := range addr.TxList {
		tx := RequestTx(t.Txid)

		if t.IsInput == true {
			for _, out := range tx.Out {
				if c[out.Address] == 0 {
					m[out.Address] = Transform{"btc.BtcAddress", "out", out.Address, LinkColor, 100, strconv.FormatFloat(out.Amount, 'f', -1, 64) + " BTC", IconURLAddr, 1}
				}
				c[out.Address]++
			}
		} else {
			for _, in := range tx.In {
				if c[in.Address] == 0 {
					m[in.Address] = Transform{"btc.BtcAddress", "in", in.Address, LinkColor, 100, strconv.FormatFloat(in.Amount, 'f', -1, 64) + " BTC", IconURLAddr, 1}
				}
				c[in.Address]++
			}
		}
	}

	// add wallet to linked entities
	Title := addr.WalletId
	if addr.Label != "" {
		Title = addr.Label
	}
	NewEnt := Transform{"btc.BtcWallet", "in", Title, LinkColor, 100, "", IconURLWt, 1}
	list.EntityList = append(list.EntityList, NewEnt)

	// add address inputs/outputs
	for _, NewEnt := range m {
		//NewEnt.Count = c[NewEnt.Value]
		list.EntityList = append(list.EntityList, NewEnt)
	}
}

func PrintTransform(list *TransformList) {
	tr := &maltegogo.MaltegoTransform{}

  for _, ent := range list.EntityList {
    NewEnt := tr.AddEntity(ent.Type, ent.Value)
		NewEnt.SetType(ent.Type)
		NewEnt.AddProperty(ent.Type, ent.Type, "stict", ent.Value)
    NewEnt.SetWeight(ent.Weight)
		NewEnt.SetLinkColor(ent.LinkColor)
    NewEnt.SetLinkLabel(ent.LinkLabel)
    NewEnt.SetIconURL(ent.IconURL)

    if ent.Direction == "in" {
      NewEnt.AddProperty("link#maltego.link.direction", "link#maltego.link.direction", "stict", "output-to-input")
    }
  }

  // print transform result
  fmt.Println(tr.ReturnOutput())
}

func InitCache() {
	var err error

	pool = zoom.NewPool(RedisURL)

	opt := zoom.DefaultCollectionOptions.WithIndex(true)
	TransformModel, err = pool.NewCollectionWithOptions(&TransformList{}, opt)
	if err != nil {
		fmt.Println(err)
	}
}

func FromCache(query string, list *TransformList) (err error) {
	// try to get from redis cache
	err = TransformModel.Find(query, list)

	if err != nil {
		return
	} else {
		log.Println("cache hit:", query)
	}

	return
}

func ClosePool() error {
	return pool.Close()
}
