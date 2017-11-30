package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"net/http"
	"io/ioutil"
)

var (
	ApiClient = "maltego-btc"
	ApiUrl = "https://www.walletexplorer.com/api/1"
)

func RequestTx(query string) (tx Tx) {
  url := ApiUrl + "/tx?txid=" + query + "&caller=" + ApiClient

	bytes := httpRequest(url)
	_ = json.Unmarshal(bytes, &tx)
	return
}

func RequestWallet(query string, from int) (wallet Wallet) {
	url := ApiUrl + "/wallet?wallet=" + query + "&from=" + strconv.Itoa(from) + "&count=100&caller=" + ApiClient

	bytes := httpRequest(url)
	_ = json.Unmarshal(bytes, &wallet)
	return
}

func RequestAddress(query string, from int) (addr Address) {
	url := ApiUrl + "/address?address=" + query + "&from=" + strconv.Itoa(from) + "&count=100&caller=" + ApiClient

	bytes := httpRequest(url)
	_ = json.Unmarshal(bytes, &addr)
	return
}

func httpRequest(url string) (bytes []byte) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("Error: " + resp.Status)
		return
	}

	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return
}
