package object

import (
	"context"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"io"
	"io/ioutil"
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
// https://github.com/fyrchik/neofs-node/blob/089f8912d277edb14b04f1d96274b792a22ed060/cmd/neofs-cli/modules/object.go#L305
func UploadObject(ctx context.Context, cli *client.Client, uploadSize int, containerID cid.ID, ownerID *owner.ID, attr []*object.Attribute, bearerToken *token.BearerToken, sessionToken *session.Token, reader *io.Reader) (oid.ID, error) {
	var objectID oid.ID
	o := object.New()
	o.SetContainerID(&containerID)
	o.SetOwnerID(ownerID)
	o.SetAttributes(attr...)

	objWriter, err := cli.ObjectPutInit(ctx, client.PrmObjectPutInit{})
	if sessionToken != nil {
		fmt.Println("using session token")
		objWriter.WithinSession(*sessionToken)
	}
	if bearerToken != nil {
		fmt.Println("using bearer token")
		objWriter.WithBearerToken(*bearerToken)
	}
	if !objWriter.WriteHeader(*o) {
		fmt.Println("error writing object header")
		return objectID, errors.New("could not write object header")
	}
	var buf []byte

	//better here to handle this based on bigger/smaller than 1024 bytes. no need to loader if smaller.
	//so instead of passing in upload type, pass in file size
	if uploadSize < 1024 {
		fmt.Println("processing small content")
		buf, err = ioutil.ReadAll(*reader)
		if err != nil {
			fmt.Println("couldn't read into buffer", err)
			return objectID, err
		}
		if !objWriter.WritePayloadChunk(buf) {
			return objectID, errors.New("couldn't write rawContent payload chunk")
		}
		fmt.Println("received ", string(buf))
	} else {
		buf = make([]byte, 1024) // 1 MiB
		for {
			// update progress bar
			_, err := (*reader).Read(buf)
			if !objWriter.WritePayloadChunk(buf) {
				break
			}
			if errors.Is(err, io.EOF) {
				break
			}
		}
	}
	res, err := objWriter.Close()
	if err != nil {
		fmt.Println("couldn't close object", err)
		return objectID, err
	}
	res.ReadStoredObjectID(&objectID)
	return objectID, err //check this might need polling to confirm success
}

func GetObjectMetaData(ctx context.Context, cli *client.Client, objectID oid.ID, containerID cid.ID, bearerToken *token.BearerToken, sessionToken *session.Token) (*object.Object, error){
	h := client.PrmObjectHead{}
	h.ByID(objectID)
	if sessionToken != nil {
		h.WithinSession(*sessionToken)
	}
	if bearerToken != nil {
		h.WithBearerToken(*bearerToken)
	}
	h.FromContainer(containerID)
	var o = &object.Object{}
	head, err := cli.ObjectHead(ctx, h)
	if err != nil {
		return o, err
	}
	response := head.ReadHeader(o)
	if !response {
		return o, errors.New("could not read the object header. Did not exist")
	}
	return o, nil
}
// GetObject does pecisely that. Returns bytes
// Todo: https://stackoverflow.com/a/56505353/1414721
// for progress bar
func GetObject(ctx context.Context, cli *client.Client, payloadSize int, objectID oid.ID, containerID cid.ID, bearerToken *token.BearerToken, sessionToken *session.Token, writer *io.Writer) (*object.Object, error){
	if writer == nil {
		return nil, errors.New("no writer provided")
	}
	dstObject := &object.Object{}
	getParms := client.PrmObjectGet{}
	getParms.ByID(objectID)
	getParms.FromContainer(containerID)
	if sessionToken != nil {
		getParms.WithinSession(*sessionToken)
	}
	if bearerToken != nil {
		getParms.WithBearerToken(*bearerToken)
	}
	objReader, err := cli.ObjectGetInit(ctx, getParms)
	if err != nil {
		return dstObject, err
	}
	if !objReader.ReadHeader(dstObject) {
		_, err = objReader.Close()
		return dstObject, err
	}
	var buf []byte
	if payloadSize < 1024 {
		buf = make([]byte, payloadSize)
		_, err := objReader.Read(buf)
		if err != nil {
			fmt.Println("couldn't read into buffer", err)
			return dstObject, err
		}
		if _, writerErr := (*writer).Write(buf); writerErr != nil {
			return nil, errors.New("error writing to buffer: " + writerErr.Error())
		}
	} else {
		buf = make([]byte, 1024)
		for {
			_, err := objReader.Read(buf)

			// get total size from object header and update progress bar based on n bytes received
			if errors.Is(err, io.EOF) {
				fmt.Println("end of file")
				break
			}
			if _, writerErr := (*writer).Write(buf); writerErr != nil {
				return nil, errors.New("error writing to buffer: " + writerErr.Error())
			}
		}
	}
	fmt.Println("finished getting")
	return dstObject, err //return pointer to avoid passing around large payloads?
}

// QueryObjects to query objects with no search terms
/*
	//var filters = object.SearchFilters{}
	//filters.AddRootFilter()
 */
func QueryObjects(ctx context.Context, cli *client.Client, containerID cid.ID, filters object.SearchFilters, bearerToken *token.BearerToken, sessionToken *session.Token) ([]oid.ID, error) {
	search := client.PrmObjectSearch{}
	if sessionToken != nil {
		search.WithinSession(*sessionToken)
	}
	if bearerToken != nil {
		search.WithBearerToken(*bearerToken)
	}
	search.SetFilters(filters)
	search.InContainer(containerID)
	
	var list []oid.ID
	searchInit, err := cli.ObjectSearchInit(ctx,search)
	if err != nil {
		return list, err
	}

	err = searchInit.Iterate(func(id oid.ID) bool {
		list = append(list, id)
		return false
	})
	return list, err
}

func DeleteObject(ctx context.Context, cli *client.Client, objectID oid.ID, containerID cid.ID, bearerToken *token.BearerToken, sessionToken *session.Token) (*client.ResObjectDelete, error) {
	del := client.PrmObjectDelete{}
	if sessionToken != nil {
		del.WithinSession(*sessionToken)
	}
	if bearerToken != nil {
		del.WithBearerToken(*bearerToken)
	}
	del.ByID(objectID)
	del.FromContainer(containerID)
	deleteResponse, err := cli.ObjectDelete(ctx, del)
	return deleteResponse, err
}
