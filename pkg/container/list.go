package container

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func List(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey) ([]*cid.ID, error) {
	// ListContainers method requires Owner ID.
	// OwnerID is a binary representation of wallet address.

	response, err := cli.ListContainers(ctx, wallet.OwnerIDFromPrivateKey(key))
	if err != nil {
		return nil, fmt.Errorf("can't list container: %w", err)
	}

	return response.IDList(), nil
}

