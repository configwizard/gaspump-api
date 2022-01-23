package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	//"github.com/amlwwalker/gaspump-api/pkg/examples/utils"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"io/ioutil"
	"log"
	"os"

	"github.com/nspcc-dev/neofs-sdk-go/client"
)

const usage = `NeoFS Balance requests

$ ./01-accounting -wallet [..] -address [..]

`

var (
	walletPath = flag.String("wallet", "", "path to JSON wallet file")
	walletAddr = flag.String("address", "", "wallet address [optional]")
	createWallet = flag.Bool("create", false, "create a wallet")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	if *createWallet {
		secureWallet, err := wallet.GenerateNewSecureWallet(*walletPath, "some account label", "password")
		if err != nil {
			log.Fatal("error generating wallet", err)
		}
		file, _ := json.MarshalIndent(secureWallet, "", " ")
		_ = ioutil.WriteFile(*walletPath, file, 0644)
		log.Printf("created new wallet\r\n%+v\r\n", file)
		os.Exit(0)
	}

	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, "password")
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(key),
		// find endpoints in https://testcdn.fs.neo.org/doc/integrations/endpoints/
		client.WithURIAddress("grpcs://st01.testnet.fs.neo.org:8082", nil),
		// check client errors in go compatible way
		client.WithNeoFSErrorParsing(),
	)
	if err != nil {
		log.Fatal("can't create NeoFS client:", err)
	}

	owner, err := wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		log.Fatal("can't get owner from private key:", err)
	}
	result, err := cli.GetBalance(ctx, owner)
	if err != nil {
		log.Fatal("can't get NeoFS Balance:", err)
	}

	fmt.Println("value:", result.Amount().Value())
	fmt.Println("precision:", result.Amount().Precision())
}


