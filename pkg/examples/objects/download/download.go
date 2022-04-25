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
	"github.com/machinebox/progress"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	obj "github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const usage = `Example

$ ./uploadObjects -wallets ../sample_wallets/wallet.json.go
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
		bearerToken, err = client2.NewBearerToken(ownerID, client2.GetHelperTokenExpiry(ctx, cli, 10), eaclTable, true, key)

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
		sessionToken, err = client2.CreateSessionWithObjectGetContext(ctx, cli, ownerID, &cntId, client2.GetHelperTokenExpiry(ctx, cli, 10), key)
		if err != nil {
			log.Fatal(err)
		}
	}
	objID := oid.ID{}
	objID.Parse(*objectID)
	head, err := object.GetObjectMetaData(ctx, cli, objID, cntId, bearerToken, sessionToken)
	if err != nil {
		log.Fatal(err)
	}
	filename := "tmp.tmp"
	for _, v := range head.Attributes() {
		if v.Key() == obj.AttributeFileName {
			filename = v.Value()
			break
		}
	}
	fmt.Printf("payload size `%+v\r\n", head.PayloadSize())
	f, err := os.Create(filepath.Join("./", filename))
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	c := progress.NewWriter(f)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		progressChan := progress.NewTicker(ctx, c, int64(head.PayloadSize()), 50*time.Millisecond)

		for p := range progressChan {
			print("time")
			fmt.Printf("\r%v remaining...", p.Remaining().Round(250*time.Millisecond))
		}
	}()
	WW := (io.Writer)(c)
	res, err := object.GetObject(ctx, cli, objID, cntId, bearerToken, sessionToken, &WW)
	if err != nil {
		log.Fatal("listing failed ", err)
	}
	wg.Wait()
	fmt.Printf("download response %+v\r\n", res)

}
//https://github.com/fyrchik/neofs-node/blob/089f8912d277edb14b04f1d96274b792a22ed060/cmd/neofs-cli/modules/object.go#L305
