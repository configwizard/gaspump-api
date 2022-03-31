package client

import (
	"context"
	"crypto/ecdsa"
	"github.com/configwizard/gaspump-api/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

func CreateSession(ctx context.Context, cli *client.Client, expiry uint64, key *ecdsa.PrivateKey) (*session.Token, error) {
	create := client.PrmSessionCreate{}
	create.SetExp(expiry)
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
	err = st.Sign(key)
	if err != nil {
		return &session.Token{}, err
	}
	return st, nil
}
func CreateSessionWithObjectGetContext(ctx context.Context, cli *client.Client, owner *owner.ID, containerID *cid.ID, expiry uint64, key *ecdsa.PrivateKey) (*session.Token, error) {
	var prmSessionCreate client.PrmSessionCreate
	prmSessionCreate.SetExp(expiry)

	stoken := session.NewToken()
	res, err := cli.SessionCreate(ctx, prmSessionCreate)
	if err != nil {
		return stoken, err
	}
	addr := address.NewAddress()
	addr.SetContainerID(containerID)

	objectCtx := session.NewObjectContext()
	objectCtx.ForGet()
	objectCtx.ApplyTo(addr)

	stoken.SetSessionKey(res.PublicKey())
	stoken.SetID(res.ID())
	stoken.SetExp(expiry)
	stoken.SetOwnerID(owner)
	stoken.SetContext(objectCtx)

	err = stoken.Sign(key)
	if err != nil {
		return stoken, err
	}
	return stoken, nil
}
func CreateSessionWithObjectDeleteContext(ctx context.Context, cli *client.Client, owner *owner.ID, containerID cid.ID, expiry uint64, key *ecdsa.PrivateKey) (*session.Token, error) {
	var prmSessionCreate client.PrmSessionCreate
	prmSessionCreate.SetExp(expiry)

	stoken := session.NewToken()
	res, err := cli.SessionCreate(ctx, prmSessionCreate)
	if err != nil {
		return stoken, err
	}
	addr := address.NewAddress()
	addr.SetContainerID(&containerID)

	objectCtx := session.NewObjectContext()
	objectCtx.ForDelete()
	objectCtx.ApplyTo(addr)

	stoken.SetSessionKey(res.PublicKey())
	stoken.SetID(res.ID())
	stoken.SetExp(expiry)
	stoken.SetOwnerID(owner)
	stoken.SetContext(objectCtx)

	err = stoken.Sign(key)
	if err != nil {
		return stoken, err
	}
	return stoken, nil
}

//alternative/reference

func CreateSessionWithContainerDeleteContext(ctx context.Context, cli *client.Client, owner *owner.ID, containerID cid.ID, expiry uint64, key *ecdsa.PrivateKey) (*session.Token, error) {
	var prmSessionCreate client.PrmSessionCreate

	prmSessionCreate.SetExp(expiry)

	res, err := cli.SessionCreate(ctx, prmSessionCreate)
	if err != nil {
		return &session.Token{}, err
	}
	addr := address.NewAddress()
	addr.SetContainerID(&containerID)


	cntContext := session.NewContainerContext()
	cntContext.IsForDelete()
	cntContext.ApplyTo(&containerID)

	objectCtx := session.NewObjectContext()
	objectCtx.ForPut()
	objectCtx.ApplyTo(addr)

	stoken := session.NewToken()
	stoken.SetSessionKey(res.PublicKey())
	stoken.SetID(res.ID())
	stoken.SetExp(expiry)
	stoken.SetOwnerID(owner)
	stoken.SetContext(cntContext)

	err = stoken.Sign(key)
	if err != nil {
		return &session.Token{}, err
	}

	return stoken, nil
}

