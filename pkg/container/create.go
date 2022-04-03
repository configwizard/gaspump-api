package container

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/configwizard/gaspump-api/pkg/wallet"

	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/policy"
)

func Create(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey, placementPolicy string, customACL acl.BasicACL, attributes []*container.Attribute) (*cid.ID, error) {
	// Put method requires Container structure.
	// Container must contain at least:
	//   - Owner
	//   - Basic ACL
	//   - Placement policy

	// Read more about placement policy syntax in specification:
	// https://github.com/nspcc-dev/neofs-spec/blob/master/01-arch/02-policy.md
	// https://github.com/nspcc-dev/neofs-api/blob/master/netmap/types.proto
	//	const placementPolicy = `REP 2 IN X
	//CBF 2
	//SELECT 2 FROM * AS X
	//`
	//	customACL := acl.EACLReadOnlyBasicRule
	containerPolicy, err := policy.Parse(placementPolicy)
	if err != nil {
		return nil, fmt.Errorf("can't parse placement policy: %w", err)
	}
	ownerID, err := wallet.OwnerIDFromPublicKey(&key.PublicKey)
	// Step 1: create container
	//containerPolicy, _ := policy.Parse("REP 2")
	cnr := container.New(
		container.WithPolicy(containerPolicy),
		container.WithOwnerID(ownerID),
		container.WithCustomBasicACL(customACL),
	)
	//cnr.SetSessionToken()
	cnr.SetAttributes(attributes)

	var prmContainerPut client.PrmContainerPut
	prmContainerPut.SetContainer(*cnr)

	cnrResponse, err := cli.ContainerPut(ctx, prmContainerPut)
	if err != nil {
		return &cid.ID{}, err
	}
	containerID := cnrResponse.ID()

	return containerID, nil
}
