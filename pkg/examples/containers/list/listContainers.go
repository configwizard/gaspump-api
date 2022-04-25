package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	container2 "github.com/configwizard/gaspump-api/pkg/container"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"io/ioutil"
	"log"
	"os"
)

const usage = `Example

$ ./createContainer -wallets ./sample_wallets/wallet.json.go
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
	useBearerToken = flag.Bool("bearer", false, "use a bearer token")

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

	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}
	w := wallet.GetWalletFromPrivateKey(key)
	log.Println("using account ", w.Address)
	cli, err := client2.NewClient(key, "grpcs://st01.testnet.fs.neo.org:8082")
	if err != nil {
		log.Fatal("can't create NeoFS client:", err)
	}

	list, err := container2.List(ctx, cli, key)

	if err != nil {
		log.Fatal("could not list containers", err)
	}

	fmt.Println("containers", len(list))
	for _, v := range list {
		fmt.Println("id ", v.String())
	}
}
