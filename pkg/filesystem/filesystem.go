package filesystem

import (
	"context"
	"crypto/ecdsa"
	client2 "github.com/amlwwalker/gaspump-api/pkg/client"
	"github.com/amlwwalker/gaspump-api/pkg/container"
	"github.com/amlwwalker/gaspump-api/pkg/object"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

type Element struct {
	ID string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"",omitempty`
	Size uint64 `json:"size"`
	Attributes map[string]string `json:"attributes""`
	Errors []error `json:"errors",omitempty`
	Children []*Element `json:"children",omitempty`
}

func GenerateFileSystemFromContainer(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey, containerID *cid.ID) Element {
	cont := Element{
		Type: "container",
		ID: containerID.String(),
		Attributes: make(map[string]string),
	}
	c, err := container.Get(ctx, cli, containerID)
	for _, a := range c.Attributes() {
		cont.Attributes[a.Key()] = a.Value()
	}
	if name, ok := cont.Attributes["name"]; ok {
		cont.Name = name
	}
	//list the contents:
	s, err := client2.CreateSession(client2.DEFAULT_EXPIRATION, ctx, cli, key)
	objs, err := object.ListObjects(ctx, cli, containerID, s)
	if err != nil {
		cont.Errors = append(cont.Errors, err)
	}
	for _, o := range objs {
		obj := Element{
			Type: "object",
			ID:         o.String(),
			Attributes: make(map[string]string),
		}
		objAddress := object.GetObjectAddress(o, containerID)
		head, err := object.GetObjectMetaData(ctx, cli, objAddress, s)
		if err != nil {
			obj.Errors = append(cont.Errors, err)
		}
		for _, a := range head.Object().Attributes() {
			obj.Attributes[a.Key()] = a.Value()
		}
		if name, ok := obj.Attributes["name"]; ok {
			obj.Name = name
		}
		obj.Size = head.Object().PayloadSize()
		cont.Size += obj.Size
		cont.Children = append(cont.Children, &obj)
	}
	return cont
}
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
