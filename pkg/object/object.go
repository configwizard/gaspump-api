package object

import (
	"context"
	"fmt"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"strconv"
)

func GetObjectAddress(objectID *object.ID, containerID *cid.ID) *object.Address {
	objAddress := object.NewAddress()
	objAddress.SetObjectID(objectID)
	objAddress.SetContainerID(containerID)
	return objAddress
}

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
func UploadObject(ctx context.Context, cli *client.Client, containerID *cid.ID, ownerID *owner.ID, attr []*object.Attribute, bearerToken *token.BearerToken, sessionToken *session.Token, reader *io.Reader) (*object.ID, error) {
	var obj = object.NewRaw()
	obj.SetContainerID(containerID)
	obj.SetOwnerID(ownerID)
	obj.SetAttributes(attr...)
	//obj.SetPayload([]byte("data"))

	//todo the headers should come in as a map so that we can search the headers and make check if these have been already set? Could leave this to the user
	// sets FileName attribute if it wasn't set from header
	//if _, ok := attr[object.AttributeFileName]; !ok {
	//	filename := object.NewAttribute()
	//	filename.SetKey(object.AttributeFileName)
	//	filename.SetValue(file.FileName())
	//	attributes = append(attributes, filename)
	//}
	//// sets Timestamp attribute if it wasn't set from header and enabled by settings
	//if _, ok := filtered[object.AttributeTimestamp]; !ok && u.enableDefaultTimestamp {
	//	timestamp := object.NewAttribute()
	//	timestamp.SetKey(object.AttributeTimestamp)
	//	timestamp.SetValue(strconv.FormatInt(time.Now().Unix(), 10))
	//	attributes = append(attributes, timestamp)
	//}

	var putParams = new(client.PutObjectParams)
	putParams.WithObject(obj.Object())
	putParams.WithPayloadReader(*reader)
	response, err := cli.PutObject(ctx, putParams, client.WithBearer(bearerToken), client.WithSession(sessionToken))
	if err != nil {
		fmt.Println("error putting object", err)
		return &object.ID{}, err
	}
	return response.ID(), err //check this might need polling to confirm success
}

func GetObjectMetaData(ctx context.Context, cli *client.Client, objectAddress *object.Address, bearerToken *token.BearerToken, sessionToken *session.Token) (*client.ObjectHeadRes, error){
	var headParams = new(client.ObjectHeaderParams)
	headParams.WithAddress(objectAddress)
	headObject, err := cli.HeadObject(ctx, headParams, client.WithBearer(bearerToken), client.WithSession(sessionToken))
	if err != nil {
		return &client.ObjectHeadRes{}, err
	}
	return headObject, nil
}
// GetObject does pecisely that. Returns bytes
// Todo: https://stackoverflow.com/a/56505353/1414721
// for progress bar
func GetObject(ctx context.Context, cli *client.Client, objectAddress *object.Address, bearerToken *token.BearerToken, sessionToken *session.Token, writer *io.Writer) ([]byte, error){
	var getParams = new(client.GetObjectParams)
	getParams.WithAddress(objectAddress)
	getParams.WithPayloadWriter(*writer)
	object, err := cli.GetObject(ctx, getParams, client.WithBearer(bearerToken), client.WithSession(sessionToken))
	if err != nil {
		return []byte{}, err
	}
	return object.Object().Payload(), nil
}

func ListObjects(ctx context.Context, cli *client.Client, containerID *cid.ID, bearerToken *token.BearerToken, sessionToken *session.Token) ([]*object.ID, error) {
	var searchParams = new (client.SearchObjectParams)
	var filters = object.SearchFilters{}
	filters.AddRootFilter()
	searchParams.WithContainerID(containerID)
	searchParams.WithSearchFilters(filters)
	res, err := cli.SearchObjects(ctx, searchParams, client.WithBearer(bearerToken), client.WithSession(sessionToken))
	if err != nil {
		return []*object.ID{}, err
	}
	return res.IDList(), nil
}

func SearchObjects(ctx context.Context, cli *client.Client, containerID *cid.ID, searchFilters object.SearchFilters, bearerToken *token.BearerToken, sessionToken *session.Token) ([]*object.ID, error) {
	var searchParams = new (client.SearchObjectParams)
	searchParams.WithContainerID(containerID)
	searchParams.WithSearchFilters(searchFilters)
	res, err := cli.SearchObjects(ctx, searchParams, client.WithBearer(bearerToken), client.WithSession(sessionToken))
	if err != nil {
		return []*object.ID{}, err
	}
	return res.IDList(), nil
}

func DeleteObject(ctx context.Context, cli *client.Client, objectAddress *object.Address, bearerToken *token.BearerToken, sessionToken *session.Token) (error) {
	var deleteParams = new (client.DeleteObjectParams)
	deleteParams.WithAddress(objectAddress)
	_, err := cli.DeleteObject(ctx, deleteParams, client.WithBearer(bearerToken), client.WithSession(sessionToken))
	return err
}
