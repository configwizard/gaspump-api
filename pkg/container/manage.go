package container

import (
	"context"
	"errors"
	"fmt"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"time"
)

func SetEACLOnContainer(ctx context.Context, cli *client.Client, containerID *cid.ID, table *eacl.Table) error {
	_, err := cli.SetEACL(ctx, table)
	if err != nil {
		return fmt.Errorf("can't set extended ACL: %w", err)
	}
	// wait for 15-30 seconds for eACL to be created
	for i := 0; i <= 30; i++ {
		if i == 30 {
			return errors.New("timeout, extended ACL was not persisted in side chain")
		}
		time.Sleep(time.Second)
		resp, err := cli.EACL(ctx, containerID)
		if err == nil {
			// there is no equal method for records yet, so we have to
			// implement it manually
			if eacl2.EqualRecords(resp.Table().Records(), table.Records()) {
				break
			}
		}
	}
	return nil
}
