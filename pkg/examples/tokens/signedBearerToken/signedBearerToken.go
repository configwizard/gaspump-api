package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"flag"
	"fmt"
	client2 "github.com/amlwwalker/gaspump-api/pkg/client"
	"github.com/amlwwalker/gaspump-api/pkg/eacl"
	"github.com/amlwwalker/gaspump-api/pkg/object"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
)

const usage = `Example

$ ./signedBearerToken -wallets ./sample_wallets/wallet.json
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
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


	gatewayKey, err := wallet.GenerateNewWallet("/tmp/tmpwallet.json")
	if err != nil {
		log.Fatal("can't creste gateway wallet:", err)
	}
	//use the gateway's key to create a client (that the gateway can then use
	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(&gatewayKey.Accounts[0].PrivateKey().PrivateKey),
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

	//ok so now we are going to have two accounts. One is the 'gateway' and one is the container owner.
	//A user will need to create a container
	//then the gateway will generate a bearer token
	//then the user will sign the bearer token data
	//then the gateway will use the bearer token to upload a file on behalf of the owner

	cntID := cid.New()
	cntID.Parse(containerID)
	rawBearerToken, bearerBytes, err := gatewayCreateToken(ctx, cli, cntID, &gatewayKey.Accounts[0].PrivateKey().PrivateKey.PublicKey)
	if err != nil {
		log.Fatal("can't create gateway token:", err)
	}
	//ok so now we have the token we want the user to sign
	signedBearerBytes, err := client2.SignBytesOnBehalf(bearerBytes, key)
	if err != nil {
		log.Fatal("can't create gateway token:", err)
	}

	ownerPublicKeyBytes := elliptic.Marshal(key.PublicKey, key.PublicKey.X, key.PublicKey.Y)
	//now the gateway recreates the bearer token
	gatewayBearerToken := client2.ReceiveSignedBearerToken(rawBearerToken, ownerPublicKeyBytes, signedBearerBytes)

	//now the gateway can act on behalf of the user, by uploading a file
	objID, err := uploadObject(ctx, cli, &key.PublicKey, cntID, filepath, gatewayBearerToken)
	if err != nil {
		log.Fatal("can't upload object on behalf of user:", err)
	}
	fmt.Printf("gateway uploaded object %s\r\n", objID)
}

func gatewayCreateToken(ctx context.Context, cli *client.Client, cid *cid.ID, key *ecdsa.PublicKey) (acl.BearerToken, []byte, error) {
	duration := int64(1)
	eaclTable, _ := eacl.AllowOthersReadOnly(cid)
	//so the tokenBytes is what we sign, but we have to reattach it to the rawToken
	return client2.GenerateUnsignedBearerToken(ctx, cli, eaclTable, duration, key)
}

func uploadObject(ctx context.Context, cli *client.Client, key *ecdsa.PublicKey, containerID *cid.ID, filepath string, bearerToken *token.BearerToken) (string, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	if err != nil {
		return "", err
	}
	//read in the data here. Note large files will hang and have to be held in memory (consider a spinner/io.Pipe)
	reader := bufio.NewReader(f)
	var ioReader io.Reader
	ioReader = reader
	ownerID, err := wallet.OwnerIDFromPublicKey(key)
	if err != nil {
		return "", err
	}
	id, err := object.UploadObject(ctx, cli, containerID, ownerID, nil, bearerToken, nil, &ioReader)
	if err != nil {
		fmt.Println("error attempting to upload", err)
	}
	return id.String(), err
}
