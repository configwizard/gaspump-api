package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	//"github.com/configwizard/gaspump-api/pkg/examples/utils"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"io/ioutil"
	"log"
	"os"

	"github.com/nspcc-dev/neofs-sdk-go/client"
)

const usage = `Example

$ ./getNetworkInfo -wallets ./sample_wallets/wallet.rawContent.go
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
)

//epoch can be useful if you want to calculate expiry time for an object, for instance
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

	info, err := client2.GetNetworkInfo(ctx, cli)
	if err != nil {
		log.Fatal("can't retrieve network info:", err)
	}
	fmt.Printf("network info %+v/r/n", info)
	msPerBlock := info.MsPerBlock()
	epoch := info.CurrentEpoch()
	fmt.Printf("current epoch %d - msPerBlock %d\n", epoch, msPerBlock)
}
