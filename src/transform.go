package main

import (
	"fmt"
	"strconv"

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

func GetTransform(query string) (list TransformList) {
	if len(query) <= 16 {
		WalletTransform(query, &list)
  } else {
		// try to get from redis cache
		err := TransformModel.Find(query, &list)
		if err != nil {
			if _, ok := err.(zoom.ModelNotFoundError); ok {
				list.Address = query
				AddressTransform(query, &list)

				// save to redis cache
				if err := TransformModel.Save(&list); err != nil {
					fmt.Println(err)
				}
			}
		}
	}

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
			for _, in := range tx.In {
				if c[in.Address] == 0 {
					m[in.Address] = Transform{"PR.BtcAddress", "out", in.Address, LinkColor, 100, "", IconURLAddr, 1}
				}
				c[in.Address]++
			}
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
					m[out.Address] = Transform{"PR.BtcAddress", "out", out.Address, LinkColor, 100, strconv.FormatFloat(out.Amount, 'f', -1, 64) + " BTC", IconURLAddr, 1}
				}
				c[out.Address]++
			}
		} else {
			for _, in := range tx.In {
				if c[in.Address] == 0 {
					m[in.Address] = Transform{"PR.BtcAddress", "in", in.Address, LinkColor, 100, strconv.FormatFloat(in.Amount, 'f', -1, 64) + " BTC", IconURLAddr, 1}
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
	NewEnt := Transform{"PR.BtcWallet", "in", Title, LinkColor, 100, "", IconURLWt, 1}
	list.EntityList = append(list.EntityList, NewEnt)

	// add address inputs/outputs
	for _, NewEnt := range m {
		//NewEnt.Count = c[NewEnt.Value]
		list.EntityList = append(list.EntityList, NewEnt)
	}
}

func TransformOut(list *TransformList) {
	tr := &MaltegoTransform{}

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

func ClosePool() error {
	return pool.Close()
}
