package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/examples/tokens/simple-share-server/api/objects"
	"github.com/configwizard/gaspump-api/pkg/examples/tokens/simple-share-server/api/tokens"
	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/policy"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const usage = `Example

$ ./uploadObjects -wallets ../sample_wallets/wallet.json
password is password
`


var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	//walletAddr = flag.String("address", "", "wallets address [optional]")
	cnt = flag.String("container", "", "choose a container")
	//createWallet = flag.Bool("create", false, "create a wallets")
	//useBearerToken = flag.Bool("bearer", false, "use a bearer token")
	password = flag.String("password", "", "wallet password")
)

func GetCredentialsFromPath(path, address, password string) (*ecdsa.PrivateKey, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read the wallets: %walletPath", err)
	}

	return getKeyFromWallet(w, address, password)
}
// getKeyFromWallet fetches private key from neo-go wallets structure
func getKeyFromWallet(w *wallet.Wallet, addrStr, password string) (*ecdsa.PrivateKey, error) {
	var (
		addr util.Uint160
		err  error
	)

	if addrStr == "" {
		addr = w.GetChangeAddress()
	} else {
		addr, err = flags.ParseAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid wallets address %s: %w", addrStr, err)
		}
	}

	acc := w.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("invalid wallets address %s: %w", addrStr, err)
	}

	if err := acc.Decrypt(password, keys.NEP2ScryptParams()); err != nil {
		return nil, errors.New("[decrypt] invalid password - " + err.Error())

	}

	return &acc.PrivateKey().PrivateKey, nil
}

func createClient(privateKey keys.PrivateKey) (*client.Client, error){
	cli, err := client.New(
		client.WithURIAddress("grpcs://st01.testnet.fs.neo.org:8082", nil),
		client.WithDefaultPrivateKey(&privateKey.PrivateKey),
		client.WithNeoFSErrorParsing(),
	)
	return cli, err
}
func createProtectedContainer(ctx context.Context, cli *client.Client, id *owner.ID) (cid.ID, error) {
	// Step 0: prepare credentials.
	// There are two keys:
	// - containerOwnerKey -- private key of the user, should be managed by wallet provider
	// - requestSenderKey -- private key of the gateway app, which will do operation on behalf of the user

	// Step 1: create container
	containerPolicy, err := policy.Parse("REP 2")
	if err != nil {
		return cid.ID{}, err
	}
	cnr := container.New(
		container.WithPolicy(containerPolicy),
		container.WithOwnerID(id),
		container.WithCustomBasicACL(acl.EACLPublicBasicRule),
	)

	var prmContainerPut client.PrmContainerPut
	prmContainerPut.SetContainer(*cnr)

	cnrResponse, err := cli.ContainerPut(ctx, prmContainerPut)
	if err != nil {
		return cid.ID{}, err
	}
	containerID := cnrResponse.ID()

	await30Seconds(func() bool {
		var prmContainerGet client.PrmContainerGet
		prmContainerGet.SetContainer(*containerID)
		_, err = cli.ContainerGet(ctx, prmContainerGet)
		fmt.Println("await error", err)
		return err == nil
	})

	fmt.Println("container ID", containerID.String())
	return *containerID, nil
}

func setRestrictedContainerAccess(ctx context.Context, cli *client.Client, containerID cid.ID) error {

	// Step 2: set restrictive extended ACL
	table := eacl2.PutAllowDenyOthersEACL(containerID, nil)//objectPutDenyOthersEACL(containerID, nil)
	var prmContainerSetEACL client.PrmContainerSetEACL
	prmContainerSetEACL.SetTable(table)

	_, err := cli.ContainerSetEACL(ctx, prmContainerSetEACL)
	if err != nil {
		return err
	}

	await30Seconds(func() bool {
		var prmContainerEACL client.PrmContainerEACL
		prmContainerEACL.SetContainer(containerID)
		r, err := cli.ContainerEACL(ctx, prmContainerEACL)
		if err != nil {
			return false
		}
		expected, err := table.Marshal()
		fmt.Println("expected marshal error ", err)
		got, err := r.Table().Marshal()
		fmt.Println("Table marshal error ", err)
		return bytes.Equal(expected, got)
	})
	return nil
}
func await30Seconds(f func() bool) {
	for i := 1; i <= 30; i++ {
		if f() {
			return
		}

		time.Sleep(time.Second)
	}
	panic("timeout")
}

func main() {

	wd, _ := os.Getwd()
	fmt.Println("pwd", wd)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	os.Setenv("PRIVATE_KEY", "1daa689d543606a7c033b7d9cd9ca793189935294f5920ef0a39b3ad0d00f190")
	// First obtain client credentials: private key of request owner
	apiPrivateKey, err := keys.NewPrivateKeyFromHex(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	//containerOwnerPrivateKey := keys.PrivateKey{PrivateKey: rawContainerPrivateKey.PrivateKey}
	//rawPublicKey, _ := containerOwnerPrivateKey.PublicKey().MarshalJSON()
	//fmt.Println("rawPublicKey ", string(rawPublicKey)) // this is the public key i am using in javascript

	apiClient, err := createClient(*apiPrivateKey)
	if err != nil {
		log.Fatal("err ", err)
	}


	//var containerID cid.ID
	if *cnt == "" {
		ctx := context.Background()
		apiKeyOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(apiPrivateKey.PublicKey()))
		//1. the container owner needs to create a container to work on:
		containerID, err := createProtectedContainer(ctx, apiClient, apiKeyOwner)
		if err != nil {
			log.Fatal("err ", err)
		}
		//2. Now the container owner needs to protect the container from undesirables
		if err := setRestrictedContainerAccess(ctx, apiClient, containerID); err != nil {
			log.Fatal("err ", err)
		}
		fmt.Println("created container id ", containerID)
	}

	// the above will have been done by the user, out of band
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to a simple example of sharing access tokens for neoFS"))
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(WalletCtx)
		//ok so this endpoint is requesting a new bearer token to sign
		r.Get("/{containerId}", tokens.UnsignedBearerToken(apiClient))
	})

	r.Route("/api/v1/object", func(r chi.Router) {
		r.Use(WalletCtx)
		r.Get("/{containerId}/{objectId}", objects.GetObjectHead(apiClient, &apiPrivateKey.PrivateKey))
	})
	http.ListenAndServe(":9000", r)
}


func WalletCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicKey := r.Header.Get("publicKey")
		stringR := r.Header.Get("X-r")
		stringS := r.Header.Get("X-s")
		ctx := context.WithValue(r.Context(), "publicKey", publicKey)
		ctx = context.WithValue(ctx, "stringR", stringR)
		ctx = context.WithValue(ctx, "stringS", stringS)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

