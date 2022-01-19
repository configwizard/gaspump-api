package container

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func Delete(ctx context.Context, cli *client.Client, containerID *cid.ID) (*client.ContainerDeleteRes, error) {
	// Delete method requires Container ID structure.
	// Container ID is walletAddr 32 byte binary value.
	// String representation of container ID encoded in Base64.

	response, err := cli.DeleteContainer(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("can't get container %s: %w", containerID, err)
	}

	return response, nil
}
