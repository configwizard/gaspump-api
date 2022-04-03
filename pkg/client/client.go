package client

import (
	"context"
	"crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
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

func GetHelperTokenExpiry(ctx context.Context, cli *client.Client, roughEpochs uint64) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return 0
	}

	expire := ni.Info().CurrentEpoch() + roughEpochs // valid for 10 epochs (~ 10 hours)
	return expire
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
func CalculateEpochsForTime(ctx context.Context, cli *client.Client, durationInSeconds int64) uint64 {
	ni, err := cli.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return 0
	}

	ms := ni.Info().MsPerBlock()
	durationInEpochs := durationInSeconds/(ms/1000) //in seconds
	return uint64(durationInEpochs) // (estimate)
}
