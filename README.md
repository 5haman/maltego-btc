## Maltego BTC

Set of Maltego transforms written in Go for Bitcoin addresses/wallets investigation. Based on [walletexplorer.com](https://www.walletexplorer.com/) API.

### Installation

Requirements:
 - Maltego 4.0 or higher
 - Go 1.8+

Installation:
- Install Blockchain.info Transform by Paterva in the Transform Hub
- Do `go install github.com/Megarushing/maltego-btc@latest`
- Download [maltego-btc.mtz] (https://github.com/Megarushing/maltego-btc/raw/master/maltego-btc.mtz)
- In Maltego go to Import | Export > Import Config
- Point to the downloaded file and import all transforms, entities and icons
- Important: Edit each Transform BTC command line to include your path to maltego-btc, this is usually `(User Folder)/go/bin/maltego-btc`

Recommended:
- I recommend also installing maltegos library standard blockchain.com transform to use alongside

### Config options

Edit config.json and re-run installation commands. List of config options:

 - ```logfile``` – path to logfile
 - ```cachefile``` – path to cache file
 - ```link_address_color``` – color of arrows from wallets and addresses
 - ```link_wallet_color``` – color of arrows from wallets to wallets
 - ```wallet_max_size``` – max count of transactions to download from api in one go
 - ```cache_addresses``` – max number of addresses to cache
 - ```cache_wallets``` – max number of wallets to cache
 - ```icon_address``` – url to address entity icon
 - ```icon_wallet``` – url to wallet entity icon
 - ```icon_service``` – url to service entity icon
 
### Screenshots

![Screenshot](assets/screenshot-1.png)
![Screenshot](assets/screenshot-2.png)
![Screenshot](assets/screenshot-3.png)

Enjoy responsibly!
