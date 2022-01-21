package client

import (
	"context"
	"crypto/ecdsa"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type FS_NETWORK string
const (
	TESTNET FS_NETWORK = "grpcs://st01.testnet.fs.neo.org:8082"
	MAINNET FS_NETWORK = "grpcs://st01.testnet.fs.neo.org:8082"
)
const DEFAULT_EXPIRATION = 140000

func NewClient(privateKey *ecdsa.PrivateKey, network FS_NETWORK) (*client.Client, error) {
	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(privateKey),
		// find endpoints in https://testcdn.fs.neo.org/doc/integrations/endpoints/
		client.WithURIAddress(string(network), nil),
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
