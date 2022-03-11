package object

import (
	"context"
	"errors"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"strconv"
)
//
//func GetObjectAddress(objectID *oid.ID, containerID *cid.ID) *oid.Address {
//	objAddress := object.NewAddress()
//	objAddress.SetObjectID(objectID)
//	objAddress.SetContainerID(containerID)
//	return objAddress
//}

// ExpireObjectByEpochAttribute there are special attributes that start with __NEOFS__
//these inform neoFS on certain specifications
//for instance __NEOFS__EXPIRATION_EPOCH expires objects at a certain date/time
//X-Attribute-Neofs-Expiration-Epoch: 100
//expiry comes from https://github.com/nspcc-dev/neofs-http-gw#uploading
//
//// ExpireObjectByDurationAttribute X-Attribute-Neofs-Expiration-Duration: 24h30m
//func ExpireObjectByDurationAttribute(duration time.Duration) (*object.Attribute) {
//	exp := object.NewAttribute()
//	exp.SetKey("__NEOFS__EXPIRATION_DURATION")
//	exp.SetValue(fmt.Sprintf("%f", duration.Seconds()))
//	return exp
//}
////ExpireObjectByTimestampAttribute X-Attribute-Neofs-Expiration-Timestamp: 1637574797
//func ExpireObjectByTimestampAttribute(expiration time.Time) (*object.Attribute) {
//	exp := object.NewAttribute()
//	exp.SetKey("__NEOFS__EXPIRATION_TIMESTAMP")
//	exp.SetValue(strconv.FormatInt(expiration.Unix(), 10))
//	return exp
//}

//ExpireObjectByTimestampAttribute X-Attribute-Neofs-Expiration-Timestamp: 1637574797
func ExpireObjectByEpochAttribute(epoch int) *object.Attribute {
	exp := object.NewAttribute()
	exp.SetKey("__NEOFS__EXPIRATION_EPOCH")
	exp.SetValue(strconv.Itoa(epoch))
	return exp
}
// UploadObject uploads from an io.Reader.
// Todo: pipe for progress https://stackoverflow.com/a/56505353/1414721
func UploadObject(ctx context.Context, cli *client.Client, containerID *cid.ID, ownerID *owner.ID, attr []*object.Attribute, bearerToken token.BearerToken, sessionToken session.Token, reader *io.Reader) (oid.ID, error) {
	var objectID oid.ID
	o := object.New()
	o.SetContainerID(containerID)
	o.SetOwnerID(ownerID)
	o.SetAttributes(attr...)

	objWriter, err := cli.ObjectPutInit(ctx, client.PrmObjectPutInit{})
	objWriter.WithinSession(sessionToken)
	objWriter.WithBearerToken(bearerToken)
	if !objWriter.WriteHeader(*o) {
		return objectID, errors.New("could not write object header")
	}
	//objWriter.Wri
	objWriter.WritePayloadChunk([]byte("data"))
	res, err := objWriter.Close()
	if err != nil {
		return objectID, err
	}

	res.ReadStoredObjectID(&objectID)

	return objectID, err //check this might need polling to confirm success
}

func GetObjectMetaData(ctx context.Context, cli *client.Client, objectID oid.ID, containerID cid.ID, bearerToken token.BearerToken, sessionToken session.Token) (object.Object, error){
	h := client.PrmObjectHead{}
	h.ByID(objectID)
	h.WithinSession(sessionToken)
	h.WithBearerToken(bearerToken)
	h.FromContainer(containerID)
	var o = object.Object{}
	head, err := cli.ObjectHead(ctx, h)
	if err != nil {
		return o, err
	}
	response := head.ReadHeader(&o)
	if !response {
		return o, errors.New("could not read the object header. Did not exist")
	}
	return o, nil
}
// GetObject does pecisely that. Returns bytes
// Todo: https://stackoverflow.com/a/56505353/1414721
// for progress bar
func GetObject(ctx context.Context, cli *client.Client, objectID oid.ID, bearerToken token.BearerToken, sessionToken session.Token, writer *io.Writer) ([]byte, error){
	getParms := client.PrmObjectGet{}
	getParms.ByID(objectID)
	getParms.WithBearerToken(bearerToken)
	getParms.WithinSession(sessionToken)
	getter, err := cli.ObjectGetInit(ctx, getParms)
	receivedBytes := []byte{}
	//getter.ReadChunk(receivedBytes)
	_, err = getter.Read(receivedBytes)
	return receivedBytes, err
}

// QueryObjects to query objects with no search terms
/*
	//var filters = object.SearchFilters{}
	//filters.AddRootFilter()
 */
func QueryObjects(ctx context.Context, cli *client.Client, containerID cid.ID, filters object.SearchFilters, bearerToken token.BearerToken, sessionToken session.Token) ([]oid.ID, error) {
	search := client.PrmObjectSearch{}
	search.WithBearerToken(bearerToken)
	search.WithinSession(sessionToken)

	search.SetFilters(filters)
	search.InContainer(containerID)
	
	var list []oid.ID
	searchInit, err := cli.ObjectSearchInit(ctx,search)
	if err != nil {
		return list, err
	}

	err = searchInit.Iterate(func(id oid.ID) bool {
		list = append(list, id)
		return true
	})
	return list, err
}

func DeleteObject(ctx context.Context, cli *client.Client, objectID oid.ID, containerID cid.ID, bearerToken token.BearerToken, sessionToken session.Token) (*client.ResObjectDelete, error) {
	del := client.PrmObjectDelete{}
	del.WithBearerToken(bearerToken)
	del.WithinSession(sessionToken)
	del.ByID(objectID)
	del.FromContainer(containerID)
	deleteResponse, err := cli.ObjectDelete(ctx, del)
	return deleteResponse, err
}
