package container

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"strconv"
	"time"

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

	ownerID, err := wallet.OwnerIDFromPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("can't retrieve owner ID: %w", err)
	}
	cnr := container.New(
		// container policy defines the way objects will be
		// placed among storage nodes from the network map
		container.WithPolicy(containerPolicy),
		// container owner can set BasicACL and remove container
		container.WithOwnerID(ownerID),
		// read more about basic ACL in specification:
		// https://github.com/nspcc-dev/neofs-spec/blob/master/01-arch/07-acl.md
		container.WithCustomBasicACL(customACL),
		// Attributes are key:value string pairs they are always optional
		container.WithAttribute(
			container.AttributeTimestamp,
			strconv.FormatInt(time.Now().Unix(), 10),
		),
	)
	cnr.SetAttributes(attributes)
	response, err := cli.PutContainer(ctx, cnr)
	if err != nil {
		return nil, fmt.Errorf("can't create new container: %w", err)
	}

	return response.ID(), nil
}
