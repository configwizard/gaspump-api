package filesystem

import (
	"context"
	"crypto/ecdsa"
	client2 "github.com/amlwwalker/gaspump-api/pkg/client"
	"github.com/amlwwalker/gaspump-api/pkg/container"
	"github.com/amlwwalker/gaspump-api/pkg/object"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type Element struct {
	ID string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"",omitempty`
	Size uint64 `json:"size"`
	Attributes map[string]string `json:"attributes""`
	Errors []error `json:"errors",omitempty`
	Children []Element `json:"children",omitempty`
}

// PopulateContainerList returns a container with its attributes as an Element (used by GenerateFileSystemFromContainer)
func PopulateContainerList(ctx context.Context, cli *client.Client, containerID *cid.ID) Element {
	cont := Element{
		Type: "container",
		ID: containerID.String(),
		Attributes: make(map[string]string),
	}
	c, err := container.Get(ctx, cli, containerID)
	c.PlacementPolicy()
	if err != nil {
		cont.Errors = append(cont.Errors, err)
		return cont
	}
	for _, a := range c.Attributes() {
		cont.Attributes[a.Key()] = a.Value()
	}
	if name, ok := cont.Attributes["name"]; ok {
		cont.Name = name
	}
	return cont
}

// GenerateFileSystemFromContainer wraps the output of GenerateObjectStruct in a container element
func GenerateFileSystemFromContainer(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey, containerID *cid.ID) Element {

	cont := PopulateContainerList(ctx, cli, containerID)
	//list the contents:
	s, err := client2.CreateSession(client2.DEFAULT_EXPIRATION, ctx, cli, key)
	objs, err := object.ListObjects(ctx, cli, containerID, s)
	if err != nil {
		cont.Errors = append(cont.Errors, err)
	}
	cont.Size, cont.Children = GenerateObjectStruct(ctx, cli, s, objs, containerID)
	return cont
}

//GenerateObjectStruct returns an array of elements containing all the objects owned by the contianer ID
func GenerateObjectStruct(ctx context.Context, cli *client.Client, s *session.Token, objs []*oid.ID, containerID *cid.ID) (uint64, []Element){
	var newObjs []Element
	size := uint64(0)
	for _, o := range objs {
		obj := Element{
			Type: "object",
			ID:         o.String(),
			Attributes: make(map[string]string),
		}
		objAddress := object.GetObjectAddress(o, containerID)
		head, err := object.GetObjectMetaData(ctx, cli, objAddress, s)
		if err != nil {
			obj.Errors = append(obj.Errors, err)
		}
		for _, a := range head.Object().Attributes() {
			obj.Attributes[a.Key()] = a.Value()
		}
		obj.Size = head.Object().PayloadSize()
		size += obj.Size
		newObjs = append(newObjs, obj)
	}
	return size, newObjs
}

//GenerateFileSystem returns an array of every object in every container the wallet key owns
func GenerateFileSystem(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey) ([]Element, error){
	var fileSystem []Element
	containerIds, err := container.List(ctx, cli, key)
	if err != nil {
		return []Element{}, err
	}
	for _, id := range containerIds {
		fileSystem = append(fileSystem, GenerateFileSystemFromContainer(ctx, cli, key, id))
	}
	return fileSystem, nil
}
