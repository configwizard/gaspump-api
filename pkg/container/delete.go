package container

import (
	"context"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neofs-sdk-go/session"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func Delete(ctx context.Context, cli *client.Client, containerID cid.ID, sessionToken *session.Token) (*client.ResContainerDelete, error) {
	// Delete method requires Container ID structure.
	// Container ID is walletAddr 32 byte binary value.
	// String representation of container ID encoded in Base64.
	containerDelete := client.PrmContainerDelete{}
	containerDelete.SetContainer(containerID)
	if sessionToken == nil {
		return &client.ResContainerDelete{}, errors.New("deleting requires a session token")
	}
	containerDelete.SetSessionToken(*sessionToken)
	cli.ContainerDelete(ctx, containerDelete)
	response, err := cli.ContainerDelete(ctx, containerDelete)
	if err != nil {
		return nil, fmt.Errorf("can't get container %s: %w", containerID, err)
	}

	return response, nil
}
