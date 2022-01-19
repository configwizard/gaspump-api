package client

import (
	"context"
	"crypto/ecdsa"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

const (
	TESTNET string = "grpcs://st01.testnet.fs.neo.org:8082"
	MAINNET = "grpcs://st01.testnet.fs.neo.org:8082"
	DEFAULT_EXPIRATION = 140000

)

func NewClient(privateKey *ecdsa.PrivateKey, network string) (*client.Client, error) {
	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(privateKey),
		// find endpoints in https://testcdn.fs.neo.org/doc/integrations/endpoints/
		client.WithURIAddress(network, nil),
		// check client errors in go compatible way
		client.WithNeoFSErrorParsing(),
	)
	return cli, err
}

func CreateSession(expiration uint64, ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey) (*session.Token, error){
	sessionResponse, err := cli.CreateSession(ctx, expiration)
	if err != nil {
		return &session.Token{}, err
	}
	st := session.NewToken()
	id, err := wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		return &session.Token{}, err
	}
	st.SetOwnerID(id)
	st.SetID(sessionResponse.ID())
	st.SetSessionKey(sessionResponse.SessionKey())
	return st, nil
}
