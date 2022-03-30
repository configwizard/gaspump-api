package tokens

import (
	"crypto/ecdsa"
	b64 "encoding/base64"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/examples/tokens/simple-share-server/api/utils"
	"github.com/go-chi/chi/v5"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"net/http"
)

func UnsignedBearerToken(cli *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		k, err, code := utils.GetPublicKey(ctx)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		cntID := cid.ID{}
		cntID.Parse(chi.URLParam(r, "containerId"))

		kOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(k))
		// Step 3. Prepare bearer token to allow PUT request
		table := eacl2.PutAllowDenyOthersEACL(cntID, k)

		//this client can be the actor's client
		bearer := token.NewBearerToken()
		bearer.SetLifetime(utils.GetHelperTokenExpiry(ctx, cli), 0, 0)
		bearer.SetEACLTable(&table)
		bearer.SetOwner(kOwner)

		//create a bearer token
		v2Bearer := bearer.ToV2()
		binaryData, _ := v2Bearer.GetBody().StableMarshal(nil)
		sEnc := b64.StdEncoding.EncodeToString(binaryData)

		w.Write([]byte(sEnc))
	}
}
