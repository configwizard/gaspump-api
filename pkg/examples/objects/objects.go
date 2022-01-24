package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	client2 "github.com/amlwwalker/gaspump-api/pkg/client"
	"github.com/amlwwalker/gaspump-api/pkg/object"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	object2 "github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"
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
	//retrieved from running the containers example
	containerID := "2qo7LZDDHJBN833dVkyDy5gwP65qBMV5uYiFMfVLjMMA"
	filepath := "./upload.gif"
	var attributes []*object2.Attribute

	sessionToken, err := client2.CreateSession(client2.DEFAULT_EXPIRATION, ctx, cli, key)
	if err != nil {
		log.Fatal(err)
	}

	cntId := new(cid.ID)
	cntId.Parse(containerID)
	objectID, err := uploadObject(ctx, cli, key, cntId, filepath, attributes, sessionToken)
	fmt.Printf("Object %s has been persisted in container %s\nview it at https://http.testnet.fs.neo.org/%s/%s", objectID, containerID, containerID, objectID)
	objectIDs, err := ListObjects(ctx, cli, cntId, sessionToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("list objects %+v\n", objectIDs)
}

func uploadObject(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey, containerID *cid.ID, filepath string, attributes []*object2.Attribute, sessionToken *session.Token) (string, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(f)
	var ioReader io.Reader
	ioReader = reader

	//set special attributes last so they don't get overwritten
	timeStampAttr := new(object2.Attribute)
	timeStampAttr.SetKey(object2.AttributeTimestamp)
	timeStampAttr.SetValue(strconv.FormatInt(time.Now().Unix(), 10))

	fileNameAttr := new(object2.Attribute)
	fileNameAttr.SetKey(object2.AttributeFileName)
	fileNameAttr.SetValue(path.Base(filepath))
	attributes = append(attributes, []*object2.Attribute{timeStampAttr, fileNameAttr}...)

	ownerID, err := wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		return "", err
	}

	id, err := object.UploadObject(ctx, cli, containerID, ownerID, attributes, sessionToken, &ioReader)
	if err != nil {
		fmt.Println("error attempting to upload", err)
	}
	return id.String(), err
}

func ListObjects(ctx context.Context, cli *client.Client, containerID *cid.ID, sessionToken *session.Token) ([]*object2.ID, error) {
	var searchParams = new (client.SearchObjectParams)
	var filters = object2.SearchFilters{}
	filters.AddRootFilter()
	searchParams.WithContainerID(containerID)
	searchParams.WithSearchFilters(filters)
	res, err := cli.SearchObjects(ctx, searchParams, client.WithSession(sessionToken))
	if err != nil {
		return []*object2.ID{}, err
	}
	return res.IDList(), nil
}
