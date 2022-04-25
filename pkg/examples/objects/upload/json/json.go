package main

import (
	"bytes"
	//"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	//"github.com/cheggaaa/pb"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/object"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	object2 "github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	//"github.com/cheggaaa/pb"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	//"sync"
	"time"
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
	password = flag.String("password", "", "wallet password")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()
	if *containerID == "" {
		log.Fatal("need a container")
	}
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
	log.Println("uploading to container", containerID)
	if *useBearerToken {
		log.Println("using bearer token...")
		sessionToken = nil
		//others:
		targetOthers := eacl.NewTarget()
		targetOthers.SetRole(eacl.RoleOthers)
		specifiedTargetRole := eacl.NewTarget()
		eacl.SetTargetECDSAKeys(specifiedTargetRole, &key.PublicKey)

		table := eacl2.PutAllowDenyOthersEACL(cntId, (*keys.PublicKey)(&key.PublicKey))
		//bearerToken, err = client2.ExampleBearerToken(30, cntId, ownerID, info.CurrentEpoch(), specifiedTargetRole, eaclTable, key)
		bearerToken, err = client2.NewBearerToken(ownerID, getHelperTokenExpiry(ctx, cli), table, true, key)

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
		sessionToken, err = client2.CreateSession(ctx, cli, client2.GetHelperTokenExpiry(ctx, cli, 10), key)
		if err != nil {
			log.Fatal(err)
		}
	}
	jsonString := []byte(`{"species": "pigeon","description": "likes to perch on rocks"}`)
	var attributes []*object2.Attribute
	objectID, err := uploadObject(ctx, cli, ownerID, cntId, jsonString, attributes, bearerToken, sessionToken)
	if err != nil {
		log.Fatal("upload failed ", err)
	}
	filter := object2.SearchFilters{}
	filter.AddRootFilter()
	fmt.Printf("Object %s has been persisted in container %s\nview it at https://http.testnet.fs.neo.org/%s/%s\r\n", objectID, *containerID, *containerID, objectID)
	_, err = object.QueryObjects(ctx, cli, cntId, filter, bearerToken, sessionToken)
	if err != nil {
		log.Fatal("listing failed ", err)
	}
}
func getHelperTokenExpiry(ctx context.Context, cli *client.Client) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		log.Fatal("error retrieving network info", err)
	}

	expire := ni.Info().CurrentEpoch() + 10 // valid for 10 epochs (~ 10 hours)
	return expire
}
func uploadObject(ctx context.Context, cli *client.Client, ownerID *owner.ID, containerID cid.ID, jsonString []byte, attributes []*object2.Attribute, bearerToken *token.BearerToken, sessionToken *session.Token) (string, error) {

	reader := bytes.NewReader(jsonString)

	RR := (io.Reader)(reader)
	//set your attributes%%%%
	timeStampAttr := new(object2.Attribute)
	timeStampAttr.SetKey(object2.AttributeTimestamp)
	timeStampAttr.SetValue(strconv.FormatInt(time.Now().Unix(), 10))

	fileNameAttr := new(object2.Attribute)
	fileNameAttr.SetKey(object2.AttributeFileName)
	fileNameAttr.SetValue("pigeons.json")
	attributes = append(attributes, []*object2.Attribute{timeStampAttr, fileNameAttr}...)

	id, err := object.UploadObject(ctx, cli, "application/json", containerID, ownerID, attributes, bearerToken, sessionToken, &RR)
	return id.String(), err
}
