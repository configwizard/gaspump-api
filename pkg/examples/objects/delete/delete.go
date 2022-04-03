package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/object"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io/ioutil"
	"log"
	"os"
)

const usage = `Example

$ ./uploadObjects -wallets ../sample_wallets/wallet.json
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	useBearerToken = flag.Bool("bearer", false, "use a bearer token")
	containerID = flag.String("container", "", "specify the container")
	objectID = flag.String("object", "", "specify the object")
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

	if *containerID == "" {
		log.Fatal("need a container")
	}
	if *objectID == "" {
		log.Fatal("need an object")
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

	cntId := cid.ID{}
	cntId.Parse(*containerID)

	ownerID := owner.NewID()
	ownerID, err = wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		log.Fatal("cant retrieve ownerID:", err)
	}
	//pointers so we can have nil tokens
	var sessionToken = &session.Token{}
	var bearerToken = &token.BearerToken{}
	log.Println("using container", containerID)
	if *useBearerToken {
		log.Println("using bearer token...")
		sessionToken = nil
		//others:
		targetOthers := eacl.NewTarget()
		targetOthers.SetRole(eacl.RoleOthers)
		specifiedTargetRole := eacl.NewTarget()
		eacl.SetTargetECDSAKeys(specifiedTargetRole, &key.PublicKey)

		eaclTable, err := eacl2.AllowKeyPutRead(cntId, specifiedTargetRole)
		if err != nil {
			log.Fatal("cant create eacl table:", err)
		}
		//(tokenReceiver *owner.ID, expire uint64, eaclTable *eacl.Table, containerOwnerKey *ecdsa.PrivateKey) (*token.BearerToken, error){
		bearerToken, err = client2.NewBearerToken(ownerID, getHelperTokenExpiry(ctx, cli), eaclTable, true, key)

		marshalBearerToken, err := client2.MarshalBearerToken(*bearerToken)
		if err != nil {
			return
		}
		fmt.Println("bearer token:", base64.StdEncoding.EncodeToString(marshalBearerToken))
		if err != nil {
			log.Fatal("cant create bearer token:", err)
		}
	} else {
		log.Println("using session token...")
		bearerToken = nil
		sessionToken, err = client2.CreateSessionWithObjectDeleteContext(ctx, cli, ownerID, cntId, client2.GetHelperTokenExpiry(ctx, cli, 10), key)
		if err != nil {
			log.Fatal(err)
		}
	}
	//Hw5z3F78HrgmCgUqw8KNkcgtaEcmv66Zxr193Nj1ZSnd


	objID := oid.ID{}
	objID.Parse(*objectID)
	//fmt.Printf("bearer %+v \r\n session %+v\r\n", bearerToken, sessionToken)
	res, err := object.DeleteObject(ctx, cli, objID, cntId, bearerToken, sessionToken)
	if err != nil {
		log.Fatal("listing failed ", err)
	}
	fmt.Printf("delete response %+v\r\n", res)
}

func getHelperTokenExpiry(ctx context.Context, cli *client.Client) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return 0
	}

	expire := ni.Info().CurrentEpoch() + 10 // valid for 10 epochs (~ 10 hours)
	return expire
}
