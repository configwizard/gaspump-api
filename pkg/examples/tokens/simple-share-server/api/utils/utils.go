package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	v2 "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"log"
	"math/big"
)

func GetHelperTokenExpiry(ctx context.Context, cli *client.Client) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		log.Fatal("cannot connect to network")
	}

	expire := ni.Info().CurrentEpoch() + 10 // valid for 10 epochs (~ 10 hours)
	return expire
}

func VerifySignature(bearer *v2.BearerToken, signatureData []byte, k keys.PublicKey) (*token.BearerToken, error) {
	v2signature := new(refs.Signature)
	v2signature.SetScheme(refs.ECDSA_SHA512)
	v2signature.SetSign(signatureData)
	v2signature.SetKey(k.Bytes())

	bearer.SetSignature(v2signature)

	newBearer := token.NewBearerTokenFromV2(bearer)
	err := newBearer.VerifySignature()
	if err != nil {
		fmt.Println("error verifying signature", err)
		return nil, errors.New("could not verify signature" + err.Error())
	}
	return newBearer, nil
}

func GetPublicKey(ctx context.Context) (*keys.PublicKey, error, int) {
	//ctx := r.Context()
	publicKey, ok := ctx.Value("publicKey").(string)
	fmt.Println("public key received", publicKey)
	if !ok {
		fmt.Println("error processing public key")
		return nil, errors.New("no valid public key"), 400
	}
	k, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		fmt.Println("error generating public key ", err)
		return nil, errors.New("no valid public key"), 400
	}
	//this should really be the actor using the bearer token
	return k, nil, 200
}

func RetriveSignatureParts(ctx context.Context) (big.Int, big.Int, error){
	stringR, ok := ctx.Value("stringR").(string)
	if !ok {
		return big.Int{}, big.Int{}, errors.New("error processing r")
	}
	stringS, ok := ctx.Value("stringS").(string)
	if !ok {
		return big.Int{}, big.Int{}, errors.New("error processing s")
	}
	r, err := bigIntByteConverter(stringR)
	if err != nil {
		return big.Int{}, big.Int{}, err
	}
	s, err := bigIntByteConverter(stringS)
	return r, s, err
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


func SignBearerTokenIfYouHavePrivateKey() {
	//h := sha512.Sum512(binaryData)
	//origR, origS, err = ecdsa.Sign(rand.Reader, &containerOwnerPrivateKey.PrivateKey, h[:])
	//if err != nil {
	//	http.Error(w, err.Error(), 400)
	//	return
	//}
	//signatureData := elliptic.Marshal(elliptic.P256(), origR, origS)
	//
	////just to prove everything is correct, create and verify the signature
	//_, err = verifySignature(v2Bearer, signatureData, *k)
	//if err != nil {
	//	http.Error(w, err.Error(), 422)
	//	return
	//}
}
