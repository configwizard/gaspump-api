package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"

	//"github.com/configwizard/gaspump-api/pkg/examples/utils"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"io/ioutil"
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
	recipient = flag.String("recipient", "NadZ8YfvkddivcFFkztZgfwxZyKf1acpRF", "wallet recipient (default NeoFS testnet")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	if *createWallet {
		secureWallet, err := wallet.GenerateNewSecureWallet(*walletPath, "some account label", *password)
		if err != nil {
			log.Fatal("error generating wallets", err)
		}
		file, _ := json.MarshalIndent(secureWallet, "", " ")
		_ = ioutil.WriteFile(*walletPath, file, 0644)
		log.Printf("created new wallets\r\n%+v\r\n", file)
		os.Exit(0)
	}

	cli, err := client.New(ctx, string(wallet.RPC_TESTNET), client.Options{})
	if err != nil {
		log.Fatal(err)
	}
	err = cli.Init()
	if err != nil {
		log.Fatal(err)
	}

	w, err := wallet.UnlockWallet(*walletPath, "", *password)
	if err != nil {
		log.Fatal("can't unlock wallet:", err)
	}
	gasToken, err := cli.GetNativeContractHash(nativenames.Gas)
	if err != nil {
		log.Fatal(err)
	}
	//send 1 GAS (precision 8) to NeoFS wallet
	//neoFSWallet := "NadZ8YfvkddivcFFkztZgfwxZyKf1acpRF"
	token, err := wallet.TransferToken(w, 1_00_000_000, *recipient, gasToken, wallet.RPC_TESTNET)
	if err != nil {
		log.Fatal("can't transfer token:", err)
	}
	log.Println("success: transaction ID ", token)
}


