package objects

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	eacl2 "github.com/configwizard/gaspump-api/pkg/eacl"
	"github.com/configwizard/gaspump-api/pkg/examples/tokens/simple-share-server/api/utils"
	"github.com/configwizard/gaspump-api/pkg/object"
	"github.com/go-chi/chi/v5"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	object2 "github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

func getBearerToken(ctx context.Context, cli *client.Client, cntID cid.ID, k *keys.PublicKey, sigR, sigS big.Int) (*token.BearerToken, error){
	kOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(k))
	signatureData := elliptic.Marshal(elliptic.P256(), &sigR, &sigS)
	table := eacl2.PutAllowDenyOthersEACL(cntID, k)

	//this client can be the actor's client
	bearer := token.NewBearerToken()
	bearer.SetLifetime(utils.GetHelperTokenExpiry(ctx, cli), 0, 0)
	bearer.SetEACLTable(&table)
	bearer.SetOwner(kOwner)

	//now sign the bearer token
	bearer, err := utils.VerifySignature(bearer.ToV2(), signatureData, *k)
	if err != nil {
		return nil, err
	}
	return bearer, nil
}
func GetObjectHead(cli *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//this is all going to get done regularly and thus should be a middleware
		cntID := cid.ID{}
		err := cntID.Parse(chi.URLParam(r, "containerId"))
		if err != nil {
			log.Println("no container id", err)
			http.Error(w, err.Error(), 400)
			return
		}
		objID := oid.ID{}
		err = objID.Parse(chi.URLParam(r, "objectId"))
		if err != nil {
			log.Println("no object id", err)
			http.Error(w, err.Error(), 400)
			return
		}
		ctx := r.Context()
		k, err, code := utils.GetPublicKey(ctx)
		if err != nil {
			log.Println("no public key", err)
			http.Error(w, err.Error(), code)
			return
		}
		sigR, sigS, err := utils.RetriveSignatureParts(ctx)
		if err != nil {
			log.Println("cannot generate signature", err)
			http.Error(w, err.Error(), 400)
			return
		}
		bearer, err := getBearerToken(ctx, cli, cntID, k, sigR, sigS)
		if err != nil {
			log.Println("cannot generate bearer token", err)
			http.Error(w, err.Error(), 400)
			return
		}

		var content *object2.Object
		content, err = object.GetObjectMetaData(ctx, cli, objID, cntID, bearer, nil)
		if err != nil {
			log.Println("cannot retrieve metadata", err)
			http.Error(w, err.Error(), 502)
			return
		}
		response, err := content.MarshalJSON()
		if err != nil {
			log.Println("cannot marhsal metadata", err)
			http.Error(w, err.Error(), 502)
			return
		}
		rEnc := b64.StdEncoding.EncodeToString(response)
		w.Header().Set("NEOFS-META", rEnc)
	}
}

func GetObject(cli *client.Client) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		cntID := cid.ID{}
		err := cntID.Parse(chi.URLParam(r, "containerId"))
		if err != nil {
			log.Println("no container id", err)
			http.Error(w, err.Error(), 400)
			return
		}
		objID := oid.ID{}
		err = objID.Parse(chi.URLParam(r, "objectId"))
		if err != nil {
			log.Println("no object id", err)
			http.Error(w, err.Error(), 400)
			return
		}
		ctx := r.Context()
		k, err, code := utils.GetPublicKey(ctx)
		if err != nil {
			log.Println("no public key", err)
			http.Error(w, err.Error(), code)
			return
		}
		sigR, sigS, err := utils.RetriveSignatureParts(ctx)
		if err != nil {
			log.Println("cannot generate signature", err)
			http.Error(w, err.Error(), 400)
			return
		}
		bearer, err := getBearerToken(ctx, cli, cntID, k, sigR, sigS)
		if err != nil {
			log.Println("cannot generate bearer token", err)
			http.Error(w, err.Error(), 400)
			return
		}
		WW := (io.Writer)(w)
		_, err = object.GetObject(ctx, cli, objID, cntID, bearer, nil, &WW)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
	}
}

func UploadObject(cli *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cntID := cid.ID{}
		err := cntID.Parse(chi.URLParam(r, "containerId"))
		if err != nil {
			log.Println("no container id", err)
			http.Error(w, err.Error(), 400)
			return
		}
		objID := oid.ID{}
		err = objID.Parse(chi.URLParam(r, "objectId"))
		if err != nil {
			log.Println("no object id", err)
			http.Error(w, err.Error(), 400)
			return
		}
		ctx := r.Context()
		k, err, code := utils.GetPublicKey(ctx)
		if err != nil {
			log.Println("no public key", err)
			http.Error(w, err.Error(), code)
			return
		}
		kOwner := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(k))
		sigR, sigS, err := utils.RetriveSignatureParts(ctx)
		if err != nil {
			log.Println("cannot generate signature", err)
			http.Error(w, err.Error(), 400)
			return
		}
		bearer, err := getBearerToken(ctx, cli, cntID, k, sigR, sigS)
		if err != nil {
			log.Println("cannot generate bearer token", err)
			http.Error(w, err.Error(), 400)
			return
		}
		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files. todo: make this larger
		r.ParseMultipartForm(10 << 20)
		// FormFile returns the first file for the given key `file`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		//e.g:
		//<form
		//enctype="multipart/form-data"
		//action="http://localhost:8080/upload"
		//method="post"
		//>
		//<input type="file" name="file" />
		//<input type="submit" value="upload" />
		//</form>

		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			http.Error(w, err.Error(), 502)
			return
		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)

		var attributes []*object2.Attribute
		//handle attributes
		parsedAttributes := make(map[string]string)
		attributesStr := r.Header.Get("NEOFS-ATTRIBUTES")
		err = json.Unmarshal([]byte(attributesStr), &parsedAttributes)
		if err != nil {
			http.Error(w, "invalid attributes" + err.Error(), 400)
			return
		}
		fmt.Printf("parsed attributes %+v\r\n", parsedAttributes)
		for k, v := range parsedAttributes {
			var tmp *object2.Attribute
			tmp.SetKey(k)
			tmp.SetKey(v)
			attributes = append(attributes, tmp)
		}
		timeStampAttr := new(object2.Attribute)
		timeStampAttr.SetKey(object2.AttributeTimestamp)
		timeStampAttr.SetValue(strconv.FormatInt(time.Now().Unix(), 10))

		fileNameAttr := new(object2.Attribute)
		fileNameAttr.SetKey(object2.AttributeFileName)
		fileNameAttr.SetValue(handler.Filename)
		attributes = append(attributes, []*object2.Attribute{timeStampAttr, fileNameAttr}...)
		RR := (io.Reader)(file)
		id, err := object.UploadObject(ctx, cli, cntID, kOwner, attributes, bearer, nil, &RR)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Write([]byte(id.String()))
	}
}
