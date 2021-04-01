package model

import (
	"fmt"
	"log"
	"strconv"

	"github.com/glennzw/maltegogo"
)

type Transform struct {
	Type      string
	Direction string
	Value     string
	Id        string
	Balance   string
	LinkColor string
	Weight    int
	LinkLabel string
	IconURL   string
	Time      uint64
}

type TransformList struct {
	Id         string
	EntityList []Transform
}

var (
	baseWeight = 10.0
)

func GetTransform(query string, Type string) (list TransformList) {
	list.Id = query
	log.Println("start transform: [", query, Type, "]")

	switch Type {
	case "WalletInfo", "WalletInOut", "WalletIn", "WalletOut":
		WalletTransform(query, &list)
	case "Wallet2Addresses":
		Wallet2AddressesTransform(query, &list)
	case "Address2Wallet":
		Address2WalletTransform(query, &list)
	default:
		log.Println("error:", "unknown transform type: "+Type)
		return
	}

	log.Println("finish transform: [", query, Type, "]")

	return
}

func FilterTransform(query string, Type string, list *TransformList) {
	list2 := []Transform{}

	for _, ent := range list.EntityList {
		switch Type {
		case "WalletInfo":
			if ent.Type == "btc.BtcWallet" && ent.Direction != "in" && ent.Direction != "out" {
				list2 = append(list2, ent)
			}
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
		case "Wallet2Addresses":
			if ent.Type == "maltego.BTCAddress" {
				list2 = append(list2, ent)
			}
		case "Address2Wallet":
			if ent.Type == "btc.BtcWallet" {
				list2 = append(list2, ent)
			}
		}
	}
	list.EntityList = list2
	return
}

func GetWeight(amount float64) (weight int) {
	weight = int(amount * baseWeight)
	if weight < 1 {
		weight = 1
	}
	return
}

func WalletTransform(query string, list *TransformList) {
	c := map[string]int{}
	m := map[string]Transform{}

	wallet := GetWallet(query)
	myLabel := wallet.WalletId
	myIcon := config.IconWallet
	myBalance := 0.0
	if len(wallet.WalletTxs) > 0 {
		myBalance = wallet.WalletTxs[0].Balance
	}
	if wallet.Label != "" {
		myLabel = wallet.Label
		myIcon = config.IconService
	}
	m[wallet.WalletId] = Transform{
		Type:    "btc.BtcWallet",
		Value:   myLabel,
		Id:      wallet.WalletId,
		Balance: strconv.FormatFloat(myBalance, 'f', -1, 64),
		IconURL: myIcon,
	}
	c[wallet.WalletId] = 1

	for _, tx := range wallet.WalletTxs {
		if tx.Type == "sent" {
			// Add outgoing links to other wallets
			for _, out := range tx.WalletTxOutputs {
				if out.WalletId != query && c[out.WalletId] == 0 {
					count := 0
					amount := 0.0
					for _, tx2 := range wallet.WalletTxs {
						for _, out2 := range tx2.WalletTxOutputs {
							if out2.WalletId == out.WalletId {
								count++
								amount += out2.Amount
							}
						}
					}
					linkLabel := strconv.FormatFloat(amount, 'f', -1, 64) + " BTC. Txs: " + strconv.Itoa(count)
					Label := out.WalletId
					Icon := config.IconWallet
					if out.Label != "" {
						Label = out.Label
						Icon = config.IconService
					}
					m[out.WalletId] = Transform{
						Type:      "btc.BtcWallet",
						Direction: "out",
						Value:     Label,
						Id:        out.WalletId,
						LinkColor: config.LinkWalletColor,
						Weight:    GetWeight(amount),
						LinkLabel: linkLabel,
						IconURL:   Icon,
						Time:      tx.Time,
					}

					c[out.WalletId] = 1
				}
			}
		} else {
			// Add incoming links from other wallets
			if c[tx.WalletId] == 0 {
				count := 0
				amount := 0.0
				for _, tx2 := range wallet.WalletTxs {
					if tx2.WalletId == tx.WalletId {
						count++
						amount += tx2.Amount
					}
				}

				Label := tx.WalletId
				Icon := config.IconWallet
				if tx.Label != "" {
					Label = tx.Label
					Icon = config.IconService
				}
				linkLabel := strconv.FormatFloat(amount, 'f', -1, 64) + " BTC. Txs: " + strconv.Itoa(count)
				m[tx.WalletId] = Transform{
					Type:      "btc.BtcWallet",
					Direction: "in",
					Value:     Label,
					Id:        tx.WalletId,
					LinkColor: config.LinkWalletColor,
					Weight:    GetWeight(amount),
					LinkLabel: linkLabel,
					IconURL:   Icon,
					Time:      tx.Time,
				}
				c[tx.WalletId] = 1
			}
		}
	}

	for _, NewEnt := range m {
		list.EntityList = append(list.EntityList, NewEnt)
	}
}

func Wallet2AddressesTransform(query string, list *TransformList) {
	c := map[string]int{}
	m := map[string]Transform{}

	addresses := GetWalletAddresses(query)

	for _, addr := range addresses.Addresses {
		m[addr.Address] = Transform{
			Type:      "maltego.BTCAddress",
			Direction: "out",
			Value:     addr.Address,
			Id:        addr.Address,
			Balance:   strconv.FormatFloat(addr.Balance, 'f', -1, 64),
			LinkColor: config.LinkAddressColor,
			LinkLabel: strconv.FormatFloat(addr.Balance, 'f', -1, 64) + " BTC",
			Weight:    GetWeight(addr.Balance),
		}
		c[addr.Address]++
	}

	// add address inputs/outputs
	for _, NewEnt := range m {
		list.EntityList = append(list.EntityList, NewEnt)
	}
}

func Address2WalletTransform(query string, list *TransformList) {
	c := map[string]int{}
	m := map[string]Transform{}

	wallet := Address2Wallet(query, 0)

	// Add incoming links from other wallets
	if c[wallet.WalletId] == 0 {
		Label := wallet.WalletId
		Icon := config.IconWallet
		if wallet.Label != "" {
			Label = wallet.Label
			Icon = config.IconService
		}
		m[wallet.WalletId] = Transform{
			Type:      "btc.BtcWallet",
			Direction: "in",
			Value:     Label,
			Id:        wallet.WalletId,
			LinkColor: config.LinkAddressColor,
			IconURL:   Icon,
		}
		c[wallet.WalletId] = 1
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
		NewEnt.AddProperty("id", "id", "stict", ent.Id)
		NewEnt.AddProperty("balance", "balance", "stict", ent.Balance)
		if ent.Weight != 0 {
			NewEnt.SetWeight(ent.Weight)
		}
		if ent.LinkColor != "" {
			NewEnt.SetLinkColor(ent.LinkColor)
		}
		NewEnt.SetLinkLabel(ent.LinkLabel)
		NewEnt.SetIconURL(ent.IconURL)

		if ent.Direction == "in" {
			NewEnt.AddProperty("link#maltego.link.direction", "link#maltego.link.direction", "stict", "output-to-input")
		}
	}

	// print transform result
	fmt.Println(tr.ReturnOutput())
}
