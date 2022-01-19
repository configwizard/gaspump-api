package client

import (
	"crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/client"
)

const (
	TESTNET string = "grpcs://st01.testnet.fs.neo.org:8082"
	MAINNET = "grpcs://st01.testnet.fs.neo.org:8082"
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
