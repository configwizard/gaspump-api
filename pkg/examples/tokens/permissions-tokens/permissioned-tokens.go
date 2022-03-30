package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"flag"
	"fmt"
	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"log"
	"os"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/policy"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"

	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
)
const usage = `Example

$ ./permissioned-tokens -wallets ../sample_wallets/wallet.json
password is password
`

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	useBearerToken = flag.Bool("bearer", false, "use a bearer token")
	password = flag.String("password", "", "wallet password")
)

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
func GetCredentialsFromPath(path, address, password string) (*ecdsa.PrivateKey, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read the wallets: %walletPath", err)
	}

	return getKeyFromWallet(w, address, password)
}
func main() {
	// Step 0: prepare credentials.
	// There are two keys:
	// - containerOwnerKey -- private key of the user, should be managed by wallet provider
	// - requestSenderKey -- private key of the gateway app, which will do operation on behalf of the user

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	// First obtain client credentials: private key of request owner
	// First obtain client credentials: private key of request owner
	rawContainerPrivateKey, err := keys.NewPrivateKeyFromHex("1daa689d543606a7c033b7d9cd9ca793189935294f5920ef0a39b3ad0d00f190")
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}
	containerOwnerPrivateKey := keys.PrivateKey{PrivateKey: rawContainerPrivateKey.PrivateKey}
	containerOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(containerOwnerPrivateKey.PublicKey()))

	requestSenderKey, _ := keys.NewPrivateKey()
	requestOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(requestSenderKey.PublicKey()))

	ctx := context.Background()

	containerOwnerClient, _ := client.New(
		client.WithURIAddress("grpcs://st01.testnet.fs.neo.org:8082", nil),
		client.WithDefaultPrivateKey(&containerOwnerPrivateKey.PrivateKey),
		client.WithNeoFSErrorParsing(),
	)

		//requestSenderClient, _ := client.New(
		//	client.WithURIAddress("grpcs://st01.testnet.fs.neo.org:8082", nil),
		//	client.WithDefaultPrivateKey(&requestSenderKey.PrivateKey),
		//	client.WithNeoFSErrorParsing(),
		//)

	// Step 1: create container
	containerPolicy, _ := policy.Parse("REP 2")
	cnr := container.New(
		container.WithPolicy(containerPolicy),
		container.WithOwnerID(containerOwner),
		container.WithCustomBasicACL(acl.EACLPublicBasicRule),
	)

	var prmContainerPut client.PrmContainerPut
	prmContainerPut.SetContainer(*cnr)

	//cnrResponse, err := containerOwnerClient.ContainerPut(ctx, prmContainerPut)
	//if err != nil {
	//	panic(err)
	//}
	//containerID := cnrResponse.ID()
	//
	//await30Seconds(func() bool {
	//	var prmContainerGet client.PrmContainerGet
	//	prmContainerGet.SetContainer(*containerID)
	//	_, err = containerOwnerClient.ContainerGet(ctx, prmContainerGet)
	//	return err == nil
	//})
	//
	//fmt.Println("container ID", containerID.String())
	//// Step 2: set restrictive extended ACL
	containerID := cid.ID{}
	containerID.Parse("FiAGxgha5YHVpKfQS26ssTdtvx7CL5YyaxJLseXSimvC")
	table := eacl2.PutAllowDenyOthersEACL(containerID, nil)//objectPutDenyOthersEACL(containerID, nil)
	//var prmContainerSetEACL client.PrmContainerSetEACL
	//prmContainerSetEACL.SetTable(table)
	//
	//_, err = containerOwnerClient.ContainerSetEACL(ctx, prmContainerSetEACL)
	//if err != nil {
	//	panic("eacl was not set")
	//}
	//
	//await30Seconds(func() bool {
	//	var prmContainerEACL client.PrmContainerEACL
	//	prmContainerEACL.SetContainer(*containerID)
	//	r, err := containerOwnerClient.ContainerEACL(ctx, prmContainerEACL)
	//	if err != nil {
	//		return false
	//	}
	//	expected, _ := table.Marshal()
	//	got, _ := r.Table().Marshal()
	//	return bytes.Equal(expected, got)
	//})

	// Step 3. Prepare bearer token to allow PUT request
	table = objectPutDenyOthersEACL(&containerID, requestSenderKey.PublicKey())

	bearer := token.NewBearerToken()
	bearer.SetLifetime(getHelperTokenExpiry(ctx, containerOwnerClient), 0, 0)
	bearer.SetEACLTable(&table)
	bearer.SetOwner(requestOwner)

	// Step 4. Sign bearer token
	// If remote signer is a program written in Go, it can use `bearer.Sign()`
	// Otherwise signer should sign stable marshalled binary message
	v2Bearer := bearer.ToV2()
	binaryData, _ := v2Bearer.GetBody().StableMarshal(nil)
	fmt.Printf("%+v\r\n", binaryData)
	h := sha512.Sum512(binaryData)
	r, s, err := ecdsa.Sign(rand.Reader, &containerOwnerPrivateKey.PrivateKey, h[:])
	if err != nil {
		panic(err)
	}
	signatureData := elliptic.Marshal(elliptic.P256(), r, s)
	fmt.Println("r-Val", r.String())
	fmt.Println("s-Val", s.String())
	fmt.Println("signature", string(signatureData))
	containerOwnerPublicKeyBytes := containerOwnerPrivateKey.PublicKey().Bytes()

	// Step 5. Receive signature and public key, set it to bearer token
	// Set signature of the bearer token by parsing it toV2 and back to the
	v2signature := new(refs.Signature)
	v2signature.SetScheme(refs.ECDSA_SHA512)
	v2signature.SetSign(signatureData)
	v2signature.SetKey(containerOwnerPublicKeyBytes)

	v2Bearer.SetSignature(v2signature)
	bearer = token.NewBearerTokenFromV2(v2Bearer)
	newBearer := token.NewBearerTokenFromV2(v2Bearer)
	err = newBearer.VerifySignature()
	if err != nil {
		fmt.Println("error verifying signature", err)
		return
	}
	// Step 5. Upload object with new bearer token
	//o := object.New()
	//o.SetContainerID(containerID)
	//o.SetOwnerID(containerOwner)
	//
	//stoken := objectSessionToken(ctx, requestSenderClient, requestOwner, containerID, &requestSenderKey.PrivateKey)
	//
	//objWriter, err := requestSenderClient.ObjectPutInit(ctx, client.PrmObjectPutInit{})
	//objWriter.WithinSession(*stoken)
	//objWriter.WithBearerToken(*bearer)
	//objWriter.WriteHeader(*o)
	//objWriter.WritePayloadChunk([]byte("Hello World"))
	//_, err = objWriter.Close()
	//if err != nil {
	//	panic(err)
	//}
}

func objectPutDenyOthersEACL(containerID *cid.ID, allowedPubKey *keys.PublicKey) eacl.Table {
	table := eacl.NewTable()
	table.SetCID(containerID)

	if allowedPubKey != nil {
		target := eacl.NewTarget()
		target.SetBinaryKeys([][]byte{allowedPubKey.Bytes()})

		allowPutRecord := eacl.NewRecord()
		allowPutRecord.SetOperation(eacl.OperationPut)
		allowPutRecord.SetAction(eacl.ActionAllow)
		allowPutRecord.SetTargets(target)

		table.AddRecord(allowPutRecord)
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

func getHelperTokenExpiry(ctx context.Context, cli *client.Client) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		panic(err)
	}

	expire := ni.Info().CurrentEpoch() + 10 // valid for 10 epochs (~ 10 hours)
	return expire
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

func objectSessionToken(ctx context.Context, cli *client.Client, owner *owner.ID, containerID *cid.ID, key *ecdsa.PrivateKey) *session.Token {
	expiry := getHelperTokenExpiry(ctx, cli)
	var prmSessionCreate client.PrmSessionCreate
	prmSessionCreate.SetExp(expiry)

	res, err := cli.SessionCreate(ctx, prmSessionCreate)
	if err != nil {
		panic(err)
	}

	addr := address.NewAddress()
	addr.SetContainerID(containerID)

	objectCtx := session.NewObjectContext()
	objectCtx.ForPut()
	objectCtx.ApplyTo(addr)

	stoken := session.NewToken()
	stoken.SetSessionKey(res.PublicKey())
	stoken.SetID(res.ID())
	stoken.SetExp(expiry)
	stoken.SetOwnerID(owner)
	stoken.SetContext(objectCtx)

	err = stoken.Sign(key)
	if err != nil {
		panic(err)
	}

	return stoken
}
