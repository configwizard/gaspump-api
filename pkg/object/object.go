package object

import (
	"context"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"io"
)

func GetObjectAddress(objectID *object.ID, containerID *cid.ID) *object.Address {
	objAddress := object.NewAddress()
	objAddress.SetObjectID(objectID)
	objAddress.SetContainerID(containerID)
	return objAddress
}
// UploadObject uploads from an io.Reader.
// Todo: pipe for progress https://stackoverflow.com/a/56505353/1414721
func UploadObject(ctx context.Context, cli *client.Client, containerID *cid.ID, ownerID *owner.ID, attr []*object.Attribute, sessionToken *session.Token, reader *io.Reader) (*object.ID, error) {
	var obj = object.NewRaw()
	obj.SetContainerID(containerID)
	obj.SetOwnerID(ownerID)
	obj.SetAttributes(attr...)
	//obj.SetPayload([]byte("data"))

	var putParams = new(client.PutObjectParams)
	putParams.WithObject(obj.Object())
	putParams.WithPayloadReader(*reader)
	response, err := cli.PutObject(ctx, putParams, client.WithSession(sessionToken))
	if err != nil {
		return &object.ID{}, err
	}
	return response.ID(), err //check this might need polling to confirm success
}

func GetObjectMetaData(ctx context.Context, cli *client.Client, objectAddress *object.Address, sessionToken *session.Token) (*client.ObjectHeadRes, error){
	var headParams = new(client.ObjectHeaderParams)
	headParams.WithAddress(objectAddress)
	headObject, err := cli.HeadObject(ctx, headParams, client.WithSession(sessionToken))
	if err != nil {
		return &client.ObjectHeadRes{}, err
	}
	return headObject, nil
}
// GetObject does pecisely that. Returns bytes
// Todo: https://stackoverflow.com/a/56505353/1414721
// for progress bar
func GetObject(ctx context.Context, cli *client.Client, objectAddress *object.Address, sessionToken *session.Token, writer *io.Writer) ([]byte, error){
	var getParams = new(client.GetObjectParams)
	getParams.WithAddress(objectAddress)
	getParams.WithPayloadWriter(*writer)
	object, err := cli.GetObject(ctx, getParams, client.WithSession(sessionToken))
	if err != nil {
		return []byte{}, err
	}
	return object.Object().Payload(), nil
}

func ListObjects(ctx context.Context, cli *client.Client, containerID *cid.ID, sessionToken *session.Token) ([]*object.ID, error) {
	var searchParams = new (client.SearchObjectParams)
	var filters = object.SearchFilters{}
	filters.AddRootFilter()
	searchParams.WithContainerID(containerID)
	searchParams.WithSearchFilters(filters)
	res, err := cli.SearchObjects(ctx, searchParams, client.WithSession(sessionToken))
	if err != nil {
		return []*object.ID{}, err
	}
	return res.IDList(), nil
}

func DeleteObject(ctx context.Context, cli *client.Client, objectID *object.Address, sessionToken *session.Token) (error) {
	var deleteParams = new (client.DeleteObjectParams)
	deleteParams.WithAddress(objectID)
	_, err := cli.DeleteObject(ctx, deleteParams, client.WithSession(sessionToken))
	return err
}
