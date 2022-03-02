package main

import (
	"bufio"
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
	object2 "github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
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
			log.Fatal("error generating wallets", err)
		}
		file, _ := json.MarshalIndent(secureWallet, "", " ")
		_ = ioutil.WriteFile(*walletPath, file, 0644)
		log.Printf("created new wallets\r\n%+v\r\n", file)
		os.Exit(0)
	}

	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, "password")
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	w := wallet.GetWalletFromPrivateKey(key)
	log.Println("using account ", w.Address)

	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(key),
		// find endpoints in https://testcdn.fs.neo.org/doc/integrations/endpoints/
		client.WithURIAddress(string(client2.TESTNET), nil),
		// check client errors in go compatible way
		client.WithNeoFSErrorParsing(),
	)
	if err != nil {
		log.Fatal("can't create NeoFS client:", err)
	}
	//retrieved from running the containers example
	containerID := "24HwJDCq7p878aLrDuu6qg3Ys3oSmxpbbmVjBLfApN8f"
	cntId := new(cid.ID)
	cntId.Parse(containerID)

	ownerID := owner.NewID()
	ownerID, err = wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		log.Fatal("cant retrieve ownerID:", err)
	}
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

		info, err := client2.GetNetworkInfo(ctx, cli)
		if err != nil {
			log.Fatal("can't get network info:", err)
		}
		eaclTable, err := eacl2.AllowKeyPutRead(cntId, specifiedTargetRole)
		if err != nil {
			log.Fatal("cant create eacl table:", err)
		}
		//bearerToken, err = client2.ExampleBearerToken(30, cntId, ownerID, info.CurrentEpoch(), specifiedTargetRole, eaclTable, key)
		bearerToken, err = client2.NewBearerToken(ctx, cli, ownerID, int64(info.CurrentEpoch() + 1000), eaclTable, key)

		marshalBearerToken, err := client2.MarshalBearerToken(bearerToken)
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
		sessionToken, err = client2.CreateSession(client2.DEFAULT_EXPIRATION, ctx, cli, key)
		if err != nil {
			log.Fatal(err)
		}
	}
	filepath := "./upload.gif"
	var attributes []*object2.Attribute
	objectID, err := uploadObject(ctx, cli, ownerID, cntId, filepath, attributes, bearerToken, sessionToken)
	if err != nil {
		log.Fatal("upload failed ", err)
	}
	fmt.Printf("Object %s has been persisted in container %s\nview it at https://http.testnet.fs.neo.org/%s/%s\r\n", objectID, containerID, containerID, objectID)
	objectIDs, err := object.ListObjects(ctx, cli, cntId, bearerToken, sessionToken)
	if err != nil {
		log.Fatal("listing failed ", err)
	}
	fmt.Printf("list objects %+v, %s\n", len(objectIDs), objectIDs[len(objectIDs) - 1])
}

func uploadObject(ctx context.Context, cli *client.Client, ownerID *owner.ID, containerID *cid.ID, filepath string, attributes []*object2.Attribute, bearerToken *token.BearerToken, sessionToken *session.Token) (string, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	if err != nil {
		return "", err
	}

	//read in the data here. Note large files will hang and have to be held in memory (consider a spinner/io.Pipe)
	reader := bufio.NewReader(f)
	var ioReader io.Reader
	ioReader = reader

	//set your attributes
	timeStampAttr := new(object2.Attribute)
	timeStampAttr.SetKey(object2.AttributeTimestamp)
	timeStampAttr.SetValue(strconv.FormatInt(time.Now().Unix(), 10))

	fileNameAttr := new(object2.Attribute)
	fileNameAttr.SetKey(object2.AttributeFileName)
	fileNameAttr.SetValue(path.Base(filepath))
	attributes = append(attributes, []*object2.Attribute{timeStampAttr, fileNameAttr}...)

	id, err := object.UploadObject(ctx, cli, containerID, ownerID, attributes, bearerToken, sessionToken, &ioReader)
	return id.String(), err
}
