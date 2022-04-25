package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"io/ioutil"
	"log"
	"os"

	"github.com/nspcc-dev/neofs-sdk-go/client"
)

const usage = `Example

$ ./retrieveNeoFSBalance -wallets ./sample_wallets/wallet.json.go
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
)

func main() {
	fmt.Println(os.Getwd())
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

	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
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
	get := client.PrmBalanceGet{}
	get.SetAccount(*owner)
	result, err := cli.BalanceGet(ctx, get)
	if err != nil {
		log.Fatal("can't get NeoFS Balance:", err)
	}

	fmt.Println("value:", result.Amount().Value())
	fmt.Println("precision:", result.Amount().Precision())
}


