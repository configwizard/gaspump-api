package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	container2 "github.com/configwizard/gaspump-api/pkg/container"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"io/ioutil"
	"log"
	"os"
)

const usage = `Example

$ ./createContainer -wallets ./sample_wallets/wallet.rawContent.go
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	containerID = flag.String("container", "", "specify container ID")
	password = flag.String("password", "", "wallet password")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *containerID == "" {
		log.Fatal("need container ID")
	}
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
	log.Println("using session token...")
	ownerID := owner.NewID()
	ownerID, err = wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		log.Fatal("cant retrieve ownerID:", err)
	}
	cntID := cid.ID{}
	cntID.Parse(*containerID)
	sessionToken, err := client2.CreateSessionWithContainerDeleteContext(ctx, cli, ownerID, cntID, client2.GetHelperTokenExpiry(ctx, cli, 10), key)
	if err != nil {
		log.Fatal(err)
	}

	//sessionToken, err := client2.CreateSession(ctx, cli, client2.GetHelperTokenExpiry(ctx, cli, 10), key)

	res, err := container2.Delete(ctx, cli, cntID, sessionToken)
	if err != nil {
		log.Fatal("could not delete containers", err)
	}
	fmt.Printf("delete response %+v\r\n", res.Status())
}
