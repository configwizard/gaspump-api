package tokens

import (
	"crypto/ecdsa"
	b64 "encoding/base64"
	client2 "github.com/configwizard/gaspump-api/pkg/client"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/examples/tokens/simple-share-server/api/utils"
	"github.com/go-chi/chi/v5"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	json "github.com/virtuald/go-ordered-json"
	"net/http"
	"time"
)
type Bearer struct {
	CreatedAt time.Time `json:"created_at"`
	Token string `json:"token"`
}
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
		table := eacl2.PutAllowDenyOthersEACL(cntID, k)
		bearer, err := client2.NewBearerToken(kOwner, utils.GetHelperTokenExpiry(ctx, cli), table, false, nil)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}

		//create a bearer token
		v2Bearer := bearer.ToV2()
		binaryData, err := v2Bearer.GetBody().StableMarshal(nil)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		sEnc := b64.StdEncoding.EncodeToString(binaryData)

		b := Bearer{
			CreatedAt: time.Now(),
			Token:     sEnc,
		}
		bEnc, err := json.Marshal(b)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		w.Write(bEnc)
	}
}
