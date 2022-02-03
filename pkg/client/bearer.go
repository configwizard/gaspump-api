package client

import (
	"context"
	"crypto/ecdsa"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)

func NewBearerToken(ctx context.Context, cli *client.Client, tokenReceiver *owner.ID, duration int64, eaclTable *eacl.Table, containerOwnerKey *ecdsa.PrivateKey) ([]byte, error){
	info, err := GetNetworkInfo(ctx, cli)
	lt := new(acl.TokenLifetime)
	lt.SetExp(CalculateEpochsForTime(info.CurrentEpoch(), duration, info.MsPerBlock())) //set the token lifetime.
	//bt.SetLifetime(lt.GetExp(), lt.GetNbf(), lt.GetIat())

	btoken := new(token.BearerToken)
	btoken.SetOwner(tokenReceiver)
	//btoken.SetLifetime(expiry, 10, 1) // exp, nbf, iat arguments like in JWT
	btoken.SetLifetime(lt.GetExp(), lt.GetNbf(), lt.GetIat()) // exp, nbf, iat arguments like in JWT
	btoken.SetEACLTable(eaclTable)

	err = btoken.SignToken(containerOwnerKey)
	if err != nil {
		return []byte{}, err
	}

	// Marshal and provide it to bearer token user
	jsonData, err := btoken.MarshalJSON()
	return jsonData, err
}
//RetrieveBearerToken returns a bearer token from the []byte array
func RetrieveBearerToken(bearerToken []byte) (*token.BearerToken, error) {
	btoken := new(token.BearerToken)
	err := btoken.UnmarshalJSON(bearerToken)
	if err != nil {
		return btoken, err
	}
	return btoken, err
}
