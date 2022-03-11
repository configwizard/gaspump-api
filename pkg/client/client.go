package client

import (
	"context"
	"crypto/ecdsa"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
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
	create := client.PrmSessionCreate{}
	create.SetExp(expiration)
	sessionResponse, err := cli.SessionCreate(ctx, create)
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
	st.SetSessionKey(sessionResponse.PublicKey())
	return st, nil
}


func GetNetworkInfo(ctx context.Context, cli *client.Client) (*netmap.NetworkInfo, error) {
	networkInfo := client.PrmNetworkInfo{}
	info, err := cli.NetworkInfo(ctx, networkInfo)
	if err != nil {
		return &netmap.NetworkInfo{}, err
	}
	return info.Info(), nil
}

// CalculateEpochsForTime takes the number of seconds into the future you want the epoch for
// and estimates it based on the current average time per epoch
func CalculateEpochsForTime(currentEpoch uint64, durationInSeconds , msPerEpoch int64) uint64 {
	//to convert a time into epochs
	//first we need to know the time per epoch
	//totalEstimatedTime := currentEpoch * timePerEpoch //(in ms)
	durationInEpochs := durationInSeconds/(msPerEpoch/1000) //in seconds
	return currentEpoch + uint64(durationInEpochs) // (estimate)
}
