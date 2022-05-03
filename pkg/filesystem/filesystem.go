package filesystem

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/configwizard/gaspump-api/pkg/container"
	"github.com/configwizard/gaspump-api/pkg/object"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	obj "github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"path/filepath"
)

type Element struct {
	ID string `json:"id"`
	Type string `josn:"type"`
	Size uint64 `json:"size"`
	BasicAcl acl.BasicACL
	ExtendedAcl eacl.Table
	Attributes map[string]string `json:"attributes""`
	Errors []error `json:"errors",omitempty`
	ParentID string
	Children []Element `json:"children",omitempty`
	PendingDeleted bool
}

// PopulateContainerList returns a container with its attributes as an Element (used by GenerateFileSystemFromContainer)
func PopulateContainerList(ctx context.Context, cli *client.Client, containerID cid.ID) Element {
	cont := Element{
		Type: "container",
		ID: containerID.String(),
		Attributes: make(map[string]string),
	}
	c, err := container.Get(ctx, cli, containerID)
	cont.BasicAcl = acl.BasicACL(c.BasicACL())
	if err != nil {
		cont.Errors = append(cont.Errors, err)
		return cont
	}
	fmt.Println("processing attributes ", c.Attributes())
	for _, a := range c.Attributes() {
		cont.Attributes[a.Key()] = a.Value()
	}
	if _, ok := cont.Attributes[obj.AttributeFileName]; !ok {
		cont.Attributes[obj.AttributeFileName] = ""
	}
	return cont
}

// GenerateFileSystemFromContainer wraps the output of GenerateObjectStruct in a container element
func GenerateFileSystemFromContainer(ctx context.Context, cli *client.Client, containerID cid.ID, bearerToken *token.BearerToken, sessionToken *session.Token) Element {
	var filters = obj.SearchFilters{}
	filters.AddRootFilter()
	cont := PopulateContainerList(ctx, cli, containerID)

	objs, err := object.QueryObjects(ctx, cli, containerID, filters, bearerToken, sessionToken)
	if err != nil {
		cont.Errors = append(cont.Errors, err)
	}
	cont.Size, cont.Children = GenerateObjectStruct(ctx, cli, objs, containerID, bearerToken, sessionToken)
	return cont
}

//GenerateObjectStruct returns an array of elements containing all the objects owned by the contianer ID
func GenerateObjectStruct(ctx context.Context, cli *client.Client, objs []oid.ID, containerID cid.ID, b *token.BearerToken, s *session.Token) (uint64, []Element){
	var newObjs []Element
	size := uint64(0)
	for _, o := range objs {
		tmp := Element{
			Type: "object",
			ID:         o.String(),
			Attributes: make(map[string]string),
		}
		//objAddress := object.GetObjectAddress(o, containerID)
		head, err := object.GetObjectMetaData(ctx, cli, o, containerID, b, s)
		if err != nil {
			tmp.Errors = append(tmp.Errors, err)
		}
		for _, a := range head.Object().Attributes() {
			tmp.Attributes[a.Key()] = a.Value()
		}
        if filename, ok := tmp.Attributes[obj.AttributeFileName]; ok {
			tmp.Attributes["X_EXT"] = filepath.Ext(filename)[1:]
        } else {
			tmp.Attributes["X_EXT"] = ""
        }

		tmp.Size = head.Object().PayloadSize()
		size += tmp.Size
		newObjs = append(newObjs, tmp)
	}
	return size, newObjs
}

//GenerateFileSystem returns an array of every object in every container the wallet key owns
func GenerateFileSystem(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey, bearerToken *token.BearerToken, sessionToken *session.Token) ([]Element, error){
	var fileSystem []Element
	containerIds, err := container.List(ctx, cli, key)
	if err != nil {
		return []Element{}, err
	}
	for _, id := range containerIds {
		fileSystem = append(fileSystem, GenerateFileSystemFromContainer(ctx, cli, *id, bearerToken, sessionToken))
	}
	return fileSystem, nil
}
