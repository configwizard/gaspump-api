package main

import (
	"context"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/acl"

	"encoding/json"
	"flag"
	"fmt"
	container2 "github.com/configwizard/gaspump-api/pkg/container"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"io/ioutil"
	"log"
	"os"
	"time"
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
	cli, err := client2.NewClient(key, client2.TESTNET)
	if err != nil {
		log.Fatal("can't create NeoFS client:", err)
	}
	var attributes []*container.Attribute
	placementPolicy := `REP 2 IN X 
	CBF 2
	SELECT 2 FROM * AS X
	`
	customACL := acl.BasicACL(acl.EACLPublicBasicRule)

	id, err := container2.Create(ctx, cli, key, placementPolicy, customACL, attributes)
	if err != nil {
		log.Fatal(err)
	}

	// Poll container ID until it will be available in the network.
	for i := 0; i <= 30; i++ {
		if i == 30 {
			log.Fatalf("Timeout, container %s was not persisted in side chain\n", id)
		}
		_, err = container2.Get(ctx, cli, *id)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	//e.g 2qo7LZDDHJBN833dVkyDy5gwP65qBMV5uYiFMfVLjMMA
	fmt.Printf("Container %s has been persisted in side chain\n", id)
}
