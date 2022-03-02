package container

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/configwizard/gaspump-api/pkg/wallet"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func List(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey) ([]*cid.ID, error) {
	// ListContainers method requires Owner ID.
	// OwnerID is a binary representation of wallets address.
	ownerID, err := wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("can't retrieve owner ID: %w", err)
	}
	response, err := cli.ListContainers(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("can't list container: %w", err)
	}

	return response.IDList(), nil
}

