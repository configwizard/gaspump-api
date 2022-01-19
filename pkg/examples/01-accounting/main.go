package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	//"github.com/nspcc-dev/neo-go/pkg/wallet"
	"amlwwalker/gas-pump/pkg/client"
	"amlwwalker/gas-pump/pkg/wallet"//"github.com/nspcc-dev/neofs-sdk-go/client"
)

const usage = `NeoFS Balance requests

$ ./01-accounting -wallet [..] -address [..]

`

var (
	walletPath = flag.String("wallet", "./wallet.json", "path to JSON wallet file")
	walletAddr = flag.String("address", "", "wallet address [optional]")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentials(*walletPath, *walletAddr)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	cli, err := client.NewClient(key, client.TESTNET)

	result, err := cli.GetBalance(ctx, wallet.OwnerIDFromPrivateKey(key))
	if err != nil {
		log.Fatal("can't get NeoFS Balance:", err)
	}

	fmt.Println("value:", result.Amount().Value())
	fmt.Println("precision:", result.Amount().Precision())
}

