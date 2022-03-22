package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	b64 "encoding/base64"
	"errors"
	"flag"
	"fmt"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/policy"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"log"
	"math/big"
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
func getHelperTokenExpiry(ctx context.Context, cli *client.Client) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		panic(err)
	}

	expire := ni.Info().CurrentEpoch() + 10 // valid for 10 epochs (~ 10 hours)
	return expire
}

func main() {

	wd, _ := os.Getwd()
	fmt.Println("pwd", wd)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	// First obtain client credentials: private key of request owner
	rawContainerPrivateKey, err := keys.NewPrivateKeyFromHex("1daa689d543606a7c033b7d9cd9ca793189935294f5920ef0a39b3ad0d00f190")
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	containerOwnerPrivateKey := keys.PrivateKey{PrivateKey: rawContainerPrivateKey.PrivateKey}
	rawPublicKey, _ := containerOwnerPrivateKey.PublicKey().MarshalJSON()
	fmt.Println("rawPublicKey ", string(rawPublicKey)) // this is the public key i am using in javascript
	containerOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(containerOwnerPrivateKey.PublicKey()))
	containerOwnerClient, err := createClient(containerOwnerPrivateKey)
	if err != nil {
		log.Fatal("err ", err)
	}
	ctx := context.Background()

	var containerID cid.ID
	if *cnt == "" {
		//1. the container owner needs to create a container to work on:
		containerID, err = createProtectedContainer(ctx, containerOwnerClient, containerOwner)
		if err != nil {
			log.Fatal("err ", err)
		}
		//2. Now the container owner needs to protect the container from undesirables
		if err := setRestrictedContainerAccess(ctx, containerOwnerClient, containerID); err != nil {
			log.Fatal("err ", err)
		}
	} else {
		fmt.Println("parsing", *cnt)
		containerID.Parse(*cnt)
	}

	// the above will have been done by the user, out of band
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to a simple example of sharing access tokens for neoFS"))
	})
	var origRString, origString string
	var origR, origS *big.Int
	r.Route("/auth/{walletAddress}", func(r chi.Router) {
		r.Use(WalletCtx)
		//ok so this endpoint is requesting a new bearer token to sign
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			publicKey, ok := ctx.Value("publicKey").(string)
			fmt.Println("public key received", publicKey)
			if !ok {
				fmt.Println("error processing public key")
				http.Error(w, http.StatusText(422), 422)
				return
			}
			k, err2 := keys.NewPublicKeyFromString(publicKey)
			if err2 != nil {
				return
			}

			pubKeyString, _ := k.MarshalJSON()
			fmt.Println("received public key ", string(pubKeyString))
			//    signature.setSign(fromByteArray(u.hexstring2ab('04' + sig.r + sig.s)));
			if err != nil {
				fmt.Println("error generating public key ", err)
				http.Error(w, http.StatusText(422), 422)
				return
			}
			//this should really be the actor using the bearer token
			requestOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(k))
			// Step 3. Prepare bearer token to allow PUT request
			table := eacl2.PutAllowDenyOthersEACL(containerID, k)

			//this client can be the actor's client
			bearer := token.NewBearerToken()
			bearer.SetLifetime(getHelperTokenExpiry(ctx, containerOwnerClient), 0, 0)
			bearer.SetEACLTable(&table)
			bearer.SetOwner(requestOwner)

			v2Bearer := bearer.ToV2()
			binaryData, _ := v2Bearer.GetBody().StableMarshal(nil)
			fmt.Printf("raw %s\r\n", binaryData)
			fmt.Println("hex")
			fmt.Println(binaryData)
			sEnc := b64.StdEncoding.EncodeToString(binaryData)

			fmt.Println("encoding ", sEnc)

			//src := "Cl8KBAgCEAsSIgog2o7q6LYowJau0IStqpK7KdMDc7a1x+Lr+LkPUvSCMzUaKQgDEAEiIxIhAvLHs6eoMwB1SpNsJFXC8et/XcCDnXNn6aakUC9BGq61GggIAxACIgIIAxIbChk1wCscYdTTROMkIzHE76olVgEs0Xdp3/biGgMIliQ="
			//orig, _ := b64.StdEncoding.DecodeString(src)
			//fmt.Println("new binary data in hex")
			//fmt.Println(orig)
			h := sha512.Sum512(binaryData)
			origR, origS, err = ecdsa.Sign(rand.Reader, &containerOwnerPrivateKey.PrivateKey, h[:])
			if err != nil {
				panic(err)
			}
			signatureData := elliptic.Marshal(elliptic.P256(), origR, origS)
			origRString = origR.String()
			origString = origS.String()
			fmt.Println("r-Val", origR.Text(16), origRString)
			fmt.Println("s-Val", origS.Text(16), origString)
			fmt.Println("signature", string(signatureData))

			v2signature := new(refs.Signature)
			v2signature.SetScheme(refs.ECDSA_SHA512)
			v2signature.SetSign(signatureData)
			v2signature.SetKey(k.Bytes())

			v2Bearer.SetSignature(v2signature)


			newBearer := token.NewBearerTokenFromV2(v2Bearer)
			err = newBearer.VerifySignature()
			if err != nil {
				fmt.Println("error verifying signature", err)
				http.Error(w, http.StatusText(422), 422)
				return
			}

			fmt.Println("sha512.Sum512(binaryData)", h)
			w.Write([]byte(sEnc))
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			stringSigR, ok := ctx.Value("stringSigR").(string)
			fmt.Println("stringSigR received", stringSigR)
			if !ok {
				fmt.Println("error processing stringSigR")
				http.Error(w, http.StatusText(422), 422)
				return
			}
			stringSigS, ok := ctx.Value("stringSigS").(string)
			fmt.Println("stringSigS received", stringSigS)
			if !ok {
				fmt.Println("error processing stringSigS")
				http.Error(w, http.StatusText(422), 422)
				return
			}
			sigR, err := bigIntByteConverter(stringSigR)
			sigS, err := bigIntByteConverter(stringSigS)
			fmt.Println(sigR)
			fmt.Println(sigS)
			//sigR := new(big.Int)
			//sigR, ok = sigR.SetString(stringSigR, 16)
			//if !ok {
			//	fmt.Println("sigR: error")
			//	http.Error(w, http.StatusText(422), 422)
			//	return
			//}
			//sigS := new(big.Int)
			//sigS, ok = sigS.SetString(stringSigS, 16)
			//if !ok {
			//	fmt.Println("sigS: error")
			//	http.Error(w, http.StatusText(422), 422)
			//	return
			//}
			publicKey, ok := ctx.Value("publicKey").(string)
			fmt.Println("public key received", publicKey)
			if !ok {
				fmt.Println("error processing public key")
				http.Error(w, http.StatusText(422), 422)
				return
			}
			k, err2 := keys.NewPublicKeyFromString(publicKey)
			if err2 != nil {
				return
			}

			pubKeyString, _ := k.MarshalJSON()
			fmt.Println("received public key ", string(pubKeyString))


			//bigIntHandler("82386781848370603388044244735974893154948281702574770963755243513728997714758", "b625441bac5635a40ede52562ddd2c86a7c7353673045e673cb50b8154a2eb46")

			signatureData := elliptic.Marshal(elliptic.P256(), &sigR, &sigS)

			//this should really be the actor using the bearer token
			requestOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(k))
			// Step 3. Prepare bearer token to allow PUT request
			table := eacl2.PutAllowDenyOthersEACL(containerID, k)

			//this client can be the actor's client
			bearer := token.NewBearerToken()
			bearer.SetLifetime(getHelperTokenExpiry(ctx, containerOwnerClient), 0, 0)
			bearer.SetEACLTable(&table)
			bearer.SetOwner(requestOwner)

			v2Bearer := bearer.ToV2()
			v2signature := new(refs.Signature)
			v2signature.SetScheme(refs.ECDSA_SHA512)
			v2signature.SetSign(signatureData)
			v2signature.SetKey(k.Bytes())

			v2Bearer.SetSignature(v2signature)
			newBearer := token.NewBearerTokenFromV2(v2Bearer)

			err = newBearer.VerifySignature()
			if err != nil {
				fmt.Println("error verifying signature", err)
				http.Error(w, http.StatusText(422), 422)
				return
			}

			marshal, err := newBearer.MarshalJSON()
			if err != nil {
				return
			}
			sEnc := b64.StdEncoding.EncodeToString(marshal)
			w.Write([]byte(sEnc))
		})
	})

	http.ListenAndServe(":9000", r)
}

func bigIntByteConverter(byteValue string) (big.Int, error) {
	bytes := new(big.Int)
	bytes, ok := bytes.SetString(byteValue, 16)
	if !ok {
		fmt.Println("by: error")
		return big.Int{}, errors.New("could not convert byteValue")
	}
	fmt.Println("recovered ", bytes)
	return *bytes, nil
}
func bigIntHandler(numValue string, byteValue string) (big.Int, big.Int, error){
	num := new(big.Int)
	num, ok := num.SetString(numValue, 10)
	if !ok {
		fmt.Println("by: error")
		return big.Int{}, big.Int{}, errors.New("could not convert numValue")
	}
	bytes := new(big.Int)
	bytes, ok = bytes.SetString(byteValue, 16)
	if !ok {
		fmt.Println("by: error")
		return big.Int{}, big.Int{}, errors.New("could not convert byteValue")
	}
	fmt.Println("num", num)
	fmt.Println("by", bytes)
	return *num, *bytes, nil
}
func WalletCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		walletAddress := chi.URLParam(r, "walletAddress")
		publicKey := r.Header.Get("publicKey")
		newSig := r.Header.Get("signature")
		stringSigR := r.Header.Get("X-r")
		stringSigS := r.Header.Get("X-s")
		ctx := context.WithValue(r.Context(), "walletAddress", walletAddress)
		ctx = context.WithValue(ctx, "publicKey", publicKey)
		ctx = context.WithValue(ctx, "stringSigR", stringSigR)
		ctx = context.WithValue(ctx, "stringSigS", stringSigS)
		ctx = context.WithValue(ctx, "signature", newSig)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

