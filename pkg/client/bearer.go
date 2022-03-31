package client

import (
	"crypto/ecdsa"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)

//duration = 30
func ExampleBearerToken(duration uint64, containerID *cid.ID, tokenReceiver *owner.ID, currentEpoch uint64, toWhom *eacl.Target, t *eacl.Table, containerOwnerKey *ecdsa.PrivateKey)(*token.BearerToken, error) {
	bt := token.NewBearerToken()
	bt.SetOwner(tokenReceiver)

	if t == nil {
		t = new(eacl.Table)
		t.SetCID(containerID)

		// order of rec is important
		rec := eacl.CreateRecord(eacl.ActionAllow, eacl.OperationPut)
		//rec.AddObjectAttributeFilter(eacl.MatchStringEqual, "Email", hashedEmail)
		//rec.SetTargets(toWhom)
		//eacl.AddFormedTarget(rec, toWhom)
		eacl.AddFormedTarget(rec, eacl.RoleOthers)

		rec2 := eacl.CreateRecord(eacl.ActionDeny, eacl.OperationGet)
		eacl.AddFormedTarget(rec2, eacl.RoleOthers)


		rec3 := eacl.CreateRecord(eacl.ActionDeny, eacl.OperationGet)
		rec3.SetTargets(toWhom)
		//eacl.AddFormedTarget(rec3, eacl.RoleOthers)
		//t.AddRecord(rec3)
		t.AddRecord(rec2)
		t.AddRecord(rec)
	}

	bt.SetEACLTable(t)

	lt := new(acl.TokenLifetime)
	lt.SetExp(currentEpoch + duration)
	bt.SetLifetime(lt.GetExp(), lt.GetNbf(), lt.GetIat())

	err := bt.SignToken(containerOwnerKey)
	return bt, err
}
func NewBearerToken(tokenReceiver *owner.ID, expire uint64, eaclTable eacl.Table, sign bool, containerOwnerKey *ecdsa.PrivateKey) (*token.BearerToken, error){
	btoken :=  token.NewBearerToken()
	btoken.SetLifetime(expire, 0, 0)
	btoken.SetOwner(tokenReceiver)
	btoken.SetEACLTable(&eaclTable)
	if sign {
		if err := btoken.SignToken(containerOwnerKey); err != nil {
			return btoken, err
		}
	}
	return btoken, nil
}
func MarshalBearerToken(btoken token.BearerToken) ([]byte, error) {
	// Marshal and provide it to bearer token user
	jsonData, err := btoken.Marshal()
	return jsonData, err
}
func MarshalBearerTokenToJson(btoken *token.BearerToken) ([]byte, error) {
	// Marshal and provide it to bearer token user
	jsonData, err := btoken.MarshalJSON()
	return jsonData, err
}
//RetrieveBearerToken returns a bearer token from the []byte array
func RetrieveBearerToken(bearerToken []byte) (*token.BearerToken, error) {
	btoken := new(token.BearerToken)
	err := btoken.UnmarshalJSON(bearerToken)
	if err != nil {
		return btoken, err
	}
	return btoken, err
}
