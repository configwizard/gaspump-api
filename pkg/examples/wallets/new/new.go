package main

import (
	"encoding/json"
	"flag"
	"fmt"
	//"github.com/configwizard/gaspump-api/pkg/examples/utils"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"log"
	"os"
)

const usage = `Example

$ ./balance -wallets ./sample_wallets/wallet.json.go
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

		secureWallet, err := wallet.GenerateNewSecureWallet(*walletPath, "some account label", *password)
		if err != nil {
			log.Fatal("error generating wallets", err)
		}
		file, _ := json.MarshalIndent(secureWallet, "", " ")
		//_ = ioutil.WriteFile(*walletPath, file, 0644)
		log.Printf("created new wallets\r\n%+v\r\n", file)

		log.Println("You should now go to the Neo Testnet Faucet and transfer your new wallet some Neo/GAS (https://neowish.ngd.network/#/)")
		os.Exit(0)
}
