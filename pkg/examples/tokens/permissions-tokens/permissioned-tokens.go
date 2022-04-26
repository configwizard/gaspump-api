package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"flag"
	"fmt"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	gpWallet "github.com/configwizard/gaspump-api/pkg/wallet"
	"log"
	"os"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/policy"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)
const usage = `Example

$ ./permissioned-tokens -wallets ../sample_wallets/wallet.rawContent.go
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	useBearerToken = flag.Bool("bearer", false, "use a bearer token")
	createContainer = flag.Bool("container", false, "create a container")
	password = flag.String("password", "", "wallet password")
)
func GetHelperTokenExpiry(ctx context.Context, cli *client.Client) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		log.Fatal("cannot connect to network")
	}

	expire := ni.Info().CurrentEpoch() + 10 // valid for 10 epochs (~ 10 hours)
	return expire
}
func main() {
	// Step 0: prepare credentials.
	// There are two keys:
	// - containerOwnerKey -- private key of the user, should be managed by wallet provider
	// - requestSenderKey -- private key of the gateway app, which will do operation on behalf of the user
	wd, _ := os.Getwd()
	fmt.Println("cwd", wd)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	//
	//if *containerID == "" {
	//	log.Fatal("need a container")
	//}
	//if *objectID == "" {
	//	log.Fatal("need an object")
	//}
	// First obtain client credentials: private key of request owner
	containerOwnerKey, err := gpWallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	ownerWallet := gpWallet.GetWalletFromPrivateKey(containerOwnerKey)
	log.Println("using account ", ownerWallet.Address)

	//cli, err := client2.NewClient(key, client2.TESTNET)
	//if err != nil {
	//	log.Fatal("can't create NeoFS client:", err)
	//}

	// Step 0: prepare credentials.
	// There are two keys:
	// - containerOwnerKey -- private key of the user, should be managed by wallet provider
	// - requestSenderKey -- private key of the gateway app, which will do operation on behalf of the user


	//containerOwnerKey, _ := keys.NewPrivateKey()
	containerOwnerID := owner.NewIDFromPublicKey(&containerOwnerKey.PublicKey)

	requestSenderKey, _ := keys.NewPrivateKey()
	requestOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(requestSenderKey.PublicKey()))

	ctx := context.Background()

	containerOwnerClient, _ := client.New(
		client.WithURIAddress(string(client2.TESTNET), nil),
		client.WithDefaultPrivateKey(containerOwnerKey),
		client.WithNeoFSErrorParsing(),
	)

	requestSenderClient, _ := client.New(
		client.WithURIAddress(string(client2.TESTNET), nil),
		client.WithDefaultPrivateKey(&requestSenderKey.PrivateKey),
		client.WithNeoFSErrorParsing(),
	)

	var containerID *cid.ID
	if *createContainer {

		// Step 1: create container
		containerPolicy, _ := policy.Parse("REP 2")
		cnr := container.New(
			container.WithPolicy(containerPolicy),
			container.WithOwnerID(containerOwnerID),
			container.WithCustomBasicACL(0x0FFFCFFF),
		)

		var prmContainerPut client.PrmContainerPut
		prmContainerPut.SetContainer(*cnr)

		cnrResponse, err := containerOwnerClient.ContainerPut(ctx, prmContainerPut)
		if err != nil {
			log.Fatalln("could not put container", err)
		}
		containerID = cnrResponse.ID()
		fmt.Println("creating container", containerID.String())
		await30Seconds(func() bool {
			var prmContainerGet client.PrmContainerGet
			prmContainerGet.SetContainer(*containerID)
			_, err = containerOwnerClient.ContainerGet(ctx, prmContainerGet)
			return err == nil
		})

		// Step 2: set restrictive extended ACL
		table := objectPutDenyOthersEACL(containerID, nil)
		var prmContainerSetEACL client.PrmContainerSetEACL
		prmContainerSetEACL.SetTable(table)

		_, err = containerOwnerClient.ContainerSetEACL(ctx, prmContainerSetEACL)
		if err != nil {
			log.Fatalln("could not set eacl", err)
		}

		await30Seconds(func() bool {
			var prmContainerEACL client.PrmContainerEACL
			prmContainerEACL.SetContainer(*containerID)
			r, err := containerOwnerClient.ContainerEACL(ctx, prmContainerEACL)
			if err != nil {
				return false
			}
			expected, _ := table.Marshal()
			got, _ := r.Table().Marshal()
			return bytes.Equal(expected, got)
		})
	} else {
		err := containerID.Parse("CodrJN9A4RpMEDKFZVSjavsYwKxwgCdBALFc1NXw3ZQg")
		if err != nil {
			log.Fatal("could not parse container ID", err)
		}
	}
	// Step 3. Prepare bearer token to allow PUT request
	table := objectPutDenyOthersEACL(containerID, requestSenderKey.PublicKey())

	bearer := token.NewBearerToken()
	bearer.SetLifetime(GetHelperTokenExpiry(ctx, requestSenderClient), 0, 0)
	bearer.SetEACLTable(&table)
	bearer.SetOwner(requestOwner)

	// Step 4. Sign bearer token
	// If remote signer is a program written in Go, it can use `bearer.Sign()`
	// Otherwise signer should sign stable marshalled binary message
	v2Bearer := bearer.ToV2()
	binaryData, _ := v2Bearer.GetBody().StableMarshal(nil)
	h := sha512.Sum512(binaryData)
	x, y, err := ecdsa.Sign(rand.Reader, containerOwnerKey, h[:])
	if err != nil {
		log.Fatalln("could not sign bearer", err)
	}
	signatureData := elliptic.Marshal(elliptic.P256(), x, y)
	containerOwnerPublicKeyBytes := ownerWallet.PrivateKey().PublicKey().Bytes()

	// Step 5. Receive signature and public key, set it to bearer token
	// Set signature of the bearer token by parsing it toV2 and back to the
	v2signature := new(refs.Signature)
	v2signature.SetScheme(refs.ECDSA_SHA512)
	v2signature.SetSign(signatureData)
	v2signature.SetKey(containerOwnerPublicKeyBytes)

	v2Bearer.SetSignature(v2signature)
	bearer = token.NewBearerTokenFromV2(v2Bearer)

	// Step 5. Upload object with new bearer token
	o := object.New()
	o.SetContainerID(containerID)
	o.SetOwnerID(containerOwnerID)

	stoken := objectSessionToken(ctx, requestSenderClient, requestOwner, containerID, &requestSenderKey.PrivateKey)

	objWriter, err := requestSenderClient.ObjectPutInit(ctx, client.PrmObjectPutInit{})
	objWriter.WithinSession(*stoken)
	objWriter.WithBearerToken(*bearer)
	objWriter.WriteHeader(*o)
	objWriter.WritePayloadChunk([]byte("Hello World"))
	_, err = objWriter.Close()
	if err != nil {
		log.Fatalln("could not put object", err)
	}
}

func objectPutDenyOthersEACL(containerID *cid.ID, allowedPubKey *keys.PublicKey) eacl.Table {
	table := eacl.NewTable()
	table.SetCID(containerID)

	if allowedPubKey != nil {
		target := eacl.NewTarget()
		target.SetBinaryKeys([][]byte{allowedPubKey.Bytes()})

		denyPutRecord := eacl.NewRecord()
		denyPutRecord.SetOperation(eacl.OperationPut)
		denyPutRecord.SetAction(eacl.ActionAllow)
		denyPutRecord.SetTargets(target)

		table.AddRecord(denyPutRecord)
	}

	target := eacl.NewTarget()
	target.SetRole(eacl.RoleOthers)

	denyPutRecord := eacl.NewRecord()
	denyPutRecord.SetOperation(eacl.OperationPut)
	denyPutRecord.SetAction(eacl.ActionDeny)
	denyPutRecord.SetTargets(target)

	table.AddRecord(denyPutRecord)

	return *table
}

func await30Seconds(f func() bool) {
	for i := 1; i <= 30; i++ {
		if f() {
			return
		}

		time.Sleep(time.Second)
	}
	log.Fatalln("await timeout")
}

func objectSessionToken(ctx context.Context, cli *client.Client, owner *owner.ID, containerID *cid.ID, key *ecdsa.PrivateKey) *session.Token {
	var prmSessionCreate client.PrmSessionCreate
	prmSessionCreate.SetExp(GetHelperTokenExpiry(ctx, cli))

	res, err := cli.SessionCreate(ctx, prmSessionCreate)
	if err != nil {
		log.Fatalln("could not create session", err)
	}

	addr := address.NewAddress()
	addr.SetContainerID(containerID)

	objectCtx := session.NewObjectContext()
	objectCtx.ForPut()
	objectCtx.ApplyTo(addr)

	stoken := session.NewToken()
	stoken.SetSessionKey(res.PublicKey())
	stoken.SetID(res.ID())
	stoken.SetExp(GetHelperTokenExpiry(ctx, cli))
	stoken.SetOwnerID(owner)
	stoken.SetContext(objectCtx)

	err = stoken.Sign(key)
	if err != nil {
		log.Fatalln("could not sign token", err)
	}

	return stoken
}
