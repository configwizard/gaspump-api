package objects

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/examples/tokens/simple-share-server/api/utils"
	"github.com/configwizard/gaspump-api/pkg/object"
	"github.com/go-chi/chi/v5"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"net/http"
)

func GetObjectHead(cli *client.Client, privateKey *ecdsa.PrivateKey) http.HandlerFunc {
return func (w http.ResponseWriter, r *http.Request) {
	//this is all going to get done regularly and thus should be a middleware
	ctx := r.Context()
	k, err, code := utils.GetPublicKey(ctx)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	kOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(k))
	objID := oid.ID{}
	err = objID.Parse(chi.URLParam(r, "objectId"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	cntID := cid.ID{}
	err = cntID.Parse(chi.URLParam(r, "containerId"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	sigR, sigS, err := utils.RetriveSignatureParts(ctx)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	signatureData := elliptic.Marshal(elliptic.P256(), &sigR, &sigS)
	table := eacl2.PutAllowDenyOthersEACL(cntID, k)

	//this client can be the actor's client
	bearer := token.NewBearerToken()
	bearer.SetLifetime(utils.GetHelperTokenExpiry(ctx, cli), 0, 0)
	bearer.SetEACLTable(&table)
	bearer.SetOwner(kOwner)

	//now sign the bearer token
	bearer, err = utils.VerifySignature(bearer.ToV2(), signatureData, *k)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	cli, err := client2.NewClient(privateKey, client2.TESTNET)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}

	head, err := object.GetObjectMetaData(ctx, cli, objID, cntID, bearer, nil)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	response, err := head.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	w.Write(response)
}
}
