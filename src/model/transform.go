package model

import (
	"fmt"
	"log"
	"strconv"

	"github.com/glennzw/maltegogo"
)

type Transform struct {
	Type					string
	Direction			string
	Value					string
	LinkColor			string
	Weight				int
	LinkLabel			string
	IconURL				string
	Time					uint64
}

type TransformList struct {
	Id						string
	EntityList		[]Transform
}

func (list *TransformList) ModelId() string {
	return list.Id
}

func (list *TransformList) SetModelId(id string) {
	list.Id = id
}

var (
	weight = 100
)

func GetTransform(query string, Type string) (list TransformList) {
	list.Id = query
	log.Println("transform:", query, Type)

	switch Type {
	case "WalletFull", "WalletInOut", "WalletIn", "WalletOut", "WalletAddr":
		WalletTransform(query, &list)
	case "AddrFull", "AddrInOut", "AddrIn", "AddrOut", "AddrWallet":
		AddressTransform(query, &list)
	default:
		log.Println("error:", "unknown transform type: " + Type)
		return
	}

	log.Println("finish transform:", query, Type)

	return
}

func FilterTransform(query string, Type string, list *TransformList) {
	list2 := []Transform{}

	for _, ent := range list.EntityList {
		switch Type {
		case "WalletFull", "AddrFull":
			list2 = append(list2, ent)
		case "WalletInOut":
			if ent.Type == "btc.BtcWallet" || ent.Type == "btc.BtcService" {
				list2 = append(list2, ent)
			}
		case "WalletIn":
			if (ent.Type == "btc.BtcWallet" || ent.Type == "btc.BtcService") && ent.Direction == "in" {
				list2 = append(list2, ent)
			}
		case "WalletOut":
			if (ent.Type == "btc.BtcWallet" || ent.Type == "btc.BtcService") && ent.Direction == "out" {
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
			if ent.Type == "btc.BtcWallet" || ent.Type == "btc.BtcService" {
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

	wallet := GetWallet(query)

	for _, tx := range wallet.TxList {
		if tx.WalletId == query {
			// Add links to wallet addresses
			for _, in := range tx.In {
				if c[in.Address] == 0 {
					weight = int(in.Amount * 100)
					m[in.Address] = Transform{"btc.BtcAddress", "out", in.Address, config.LinkColor1, weight, strconv.FormatFloat(in.Amount, 'f', -1, 64) + " BTC", config.IconAddress, tx.Time}
				}
				c[in.Address]++
			}

			// Add links to other wallets
			for _, out := range tx.Out {
				if out.WalletId != query && c[out.WalletId] == 0 {
					wallet2 := GetWallet(out.WalletId)
					count := 0
					amount := 0.0
					for _, tx2 := range wallet.TxList {
						for _, out2 := range tx2.Out {
							if out2.WalletId == wallet2.WalletId {
								count++
								amount += out2.Amount
							}
						}
					}
					weight = int(amount * 100)
					linkLabel := "Total: " + strconv.FormatFloat(amount, 'f', -1, 64) + " BTC. Txs: " + strconv.Itoa(count)

					Label := wallet2.WalletId
					if wallet2.Label != "" {
						Label = wallet2.Label
					}

					if wallet2.TxCount > config.TxsThreshold {
						m[out.WalletId] = Transform{"btc.BtcService", "out", Label, config.LinkColor2, weight, linkLabel, config.IconService, tx.Time}
					} else {
						m[out.WalletId] = Transform{"btc.BtcWallet", "out", Label, config.LinkColor1, weight, linkLabel, config.IconWallet, tx.Time}
					}
					c[out.WalletId] = 1
				}
			}
		} else {
			// Add incoming links to other wallets
			if c[tx.WalletId] == 0 {
				wallet2 := GetWallet(tx.WalletId)
				count := 0
				amount := 0.0
				for _, tx2 := range wallet2.TxList {
					for _, out2 := range tx2.Out {
						if out2.WalletId == query {
							count++
							amount += out2.Amount
						}
					}
				}
				Label := wallet2.WalletId
				if wallet2.Label != "" {
					Label = wallet2.Label
				}
				weight = int(amount * 100)
				linkLabel := "Total: " + strconv.FormatFloat(amount, 'f', -1, 64) + " BTC. Txs: " + strconv.Itoa(count)
				if wallet2.TxCount > config.TxsThreshold {
					m[wallet.WalletId] = Transform{"btc.BtcService", "in", Label, config.LinkColor2, weight, linkLabel, config.IconService, tx.Time}
				} else {
					m[wallet.WalletId] = Transform{"btc.BtcWallet", "in", Label, config.LinkColor1, weight, linkLabel, config.IconWallet, tx.Time}
				}
				c[tx.WalletId] = 1
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

	addr := GetAddress(query)

	for _, tx := range addr.TxList {
		if tx.WalletId == query {
			for _, out := range tx.Out {
				if c[out.Address] == 0 {
					weight = int(out.Amount * 100)
					m[out.Address] = Transform{"btc.BtcAddress", "out", out.Address, config.LinkColor1, weight, strconv.FormatFloat(out.Amount, 'f', -1, 64) + " BTC", config.IconAddress, tx.Time}
				}
				c[out.Address]++
			}
		} else {
			for _, in := range tx.In {
				if c[in.Address] == 0 {
					weight = int(in.Amount * 100)
					m[in.Address] = Transform{"btc.BtcAddress", "in", in.Address, config.LinkColor2, weight, strconv.FormatFloat(in.Amount, 'f', -1, 64) + " BTC", config.IconAddress, tx.Time}
				}
				c[in.Address]++
			}
		}
	}

	// add wallet to linked entities
	wallet := GetWallet(addr.WalletId)

	Label := wallet.WalletId
	if wallet.Label != "" {
		Label = wallet.Label
	}

	// check for large services wallets
	weight = 100
	if wallet.TxCount > config.TxsThreshold {
		NewEnt := Transform{"btc.BtcService", "in", Label, config.LinkColor2, weight, "", config.IconService, 0}
		list.EntityList = append(list.EntityList, NewEnt)
	} else {
		NewEnt := Transform{"btc.BtcWallet", "in", Label, config.LinkColor1, weight, "", config.IconWallet, 0}
		list.EntityList = append(list.EntityList, NewEnt)
	}

	// add address inputs/outputs
	for _, NewEnt := range m {
		list.EntityList = append(list.EntityList, NewEnt)
	}
}

func PrintTransform(list *TransformList) {
	tr := &maltegogo.MaltegoTransform{}

	// generate transform result
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
