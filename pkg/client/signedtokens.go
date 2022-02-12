package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	session2 "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)


func GenerateUnsignedSessionToken(ctx context.Context, cli *client.Client, expiration uint64, table *eacl.Table, duration int64, authorizedPublicKey *ecdsa.PublicKey) (session2.SessionToken, []byte, error) {
	sessionResponse, err := cli.CreateSession(ctx, expiration)
	if err != nil {
		return session2.SessionToken{}, []byte{}, err
	}
	st := session.NewToken()
	id, err := wallet.OwnerIDFromPublicKey(authorizedPublicKey)
	if err != nil {
		return session2.SessionToken{}, []byte{}, err
	}
	st.SetOwnerID(id)
	st.SetID(sessionResponse.ID())
	st.SetSessionKey(sessionResponse.SessionKey())

	rawSessionToken := st.ToV2()
	binaryData, err := rawSessionToken.GetBody().StableMarshal(nil)
	return *rawSessionToken, binaryData, err
}
//giving access to a public key
func GenerateUnsignedBearerToken(ctx context.Context, cli *client.Client, table *eacl.Table, duration int64, authorizedPublicKey *ecdsa.PublicKey) (acl.BearerToken, []byte, error) {
	// Create this bearer token on RESTFul API Gateway side.
	//var APIGatewayPublicKey *ecdsa.PublicKey
	oid, err := wallet.OwnerIDFromPublicKey(authorizedPublicKey)
	if err != nil {
		return acl.BearerToken{}, []byte{}, err
	}
	info, err := GetNetworkInfo(ctx, cli)
	if err != nil {
		return acl.BearerToken{}, []byte{}, err
	}
	lt := new(acl.TokenLifetime)
	lt.SetExp(CalculateEpochsForTime(info.CurrentEpoch(), duration, info.MsPerBlock())) //set the token lifetime

	bearerToken := token.NewBearerToken()
	bearerToken.SetEACLTable(table)
	bearerToken.SetOwner(oid)
	bearerToken.SetLifetime(lt.GetExp(), lt.GetNbf(), lt.GetIat())
	// SDK does not provide function to sign token other than `signedBearerToken.SignToken`.
	// We want to sign the structure by sending its data to the external wallet provider.
	// It is possible to do by converting SDK structure to RAW protobuf structure without syntax sugar.
	rawBearerToken := bearerToken.ToV2()
	// Marshal token body into binary.
	binaryData, err := rawBearerToken.GetBody().StableMarshal(nil)
	// Send this binary data to wallet provider to sign
	return *rawBearerToken, binaryData, err
}

// must be done by the container owner
func SignBytesOnBehalf(binaryData []byte, privateKey *ecdsa.PrivateKey) ([]byte, error ){

	// this sign procedure is copy-pasted from
	// https://github.com/nspcc-dev/neofs-sdk-go/blob/40aaaafc73a6b90583b373f231b6e8f0523bf59f/util/signature/options.go#L28
	h := sha512.Sum512(binaryData)
	x, y, err := ecdsa.Sign(rand.Reader, privateKey, h[:])
	if err != nil {
		return nil, err
	}
	//the signed bytes
	return elliptic.Marshal(elliptic.P256(), x, y), nil
}
// ReceiveSignedBearerToken takes the raw signed token and reattaches it to a token that the gateway can then use
// 	ownerPublicKey is the 33 byte public key from wallet provider
//	publicKey := elliptic.Marshal(ownerPublicKey, ownerPublicKey.X, ownerPublicKey.Y)
func ReceiveSignedBearerToken(rawBearerToken acl.BearerToken, ownerPublicKey []byte, signatureData []byte) *token.BearerToken {
	// Attach signature to the bearer token,
	signature := new(refs.Signature) // RAW protobuf structure github.com/nspcc-dev/neofs-api-go/v2/refs
	signature.SetSign(signatureData)
	signature.SetKey(ownerPublicKey) //this is the actual owner, not the gateway
	rawBearerToken.SetSignature(signature)
	// We can convert RAW protobuf structure back to SDK structure.
	signedBearerToken := token.NewBearerTokenFromV2(&rawBearerToken)
	return signedBearerToken
}

// ReceiveSignedSessionToken takes the raw signed token and reattaches it to a token that the gateway can then use
// 	ownerPublicKey is the 33 byte public key from wallet provider
//	publicKey := elliptic.Marshal(ownerPublicKey, ownerPublicKey.X, ownerPublicKey.Y)
func ReceiveSignedSessionToken(rawSessionToken session2.SessionToken, ownerPublicKey []byte, signatureData []byte) *session.Token {

	// Attach signature to the bearer token,
	signature := new(refs.Signature) // RAW protobuf structure github.com/nspcc-dev/neofs-api-go/v2/refs
	signature.SetSign(signatureData)
	signature.SetKey(ownerPublicKey) //this is the actual owner, not the gateway
	rawSessionToken.SetSignature(signature)
	signedSessionToken := session.NewTokenFromV2(&rawSessionToken)
	// We can convert RAW protobuf structure back to SDK structure.
	return signedSessionToken
}
