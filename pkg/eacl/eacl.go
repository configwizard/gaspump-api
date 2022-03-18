package eacl

import (
	"bytes"
	"crypto/ecdsa"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
)

func PutAllowDenyOthersEACL(containerID *cid.ID, allowedPubKey *keys.PublicKey) eacl.Table {
	table := eacl.NewTable()
	table.SetCID(containerID)

	if allowedPubKey != nil {
		target := eacl.NewTarget()
		target.SetBinaryKeys([][]byte{allowedPubKey.Bytes()})

		allowPutRecord := eacl.NewRecord()
		allowPutRecord.SetOperation(eacl.OperationPut)
		allowPutRecord.SetAction(eacl.ActionAllow)
		allowPutRecord.SetTargets(target)

		table.AddRecord(allowPutRecord)
	}

	target := eacl.NewTarget()
	target.SetRole(eacl.RoleOthers)

	denyPutRecord := eacl.NewRecord()
	denyPutRecord.SetOperation(eacl.OperationPut)
	denyPutRecord.SetAction(eacl.ActionDeny)
	denyPutRecord.SetTargets(target)

	table.AddRecord(denyPutRecord)

	return *table
}

//AllowOthersReadOnly from https://github.com/nspcc-dev/neofs-s3-gw/blob/fdc07b8dc15272e2aabcbd7bb8c19e435c94e392/authmate/authmate.go#L358
func AllowKeyPutRead(cid *cid.ID, toWhom *eacl.Target) (*eacl.Table, error) {
	table := eacl.NewTable()
	targetOthers := eacl.NewTarget()
	targetOthers.SetRole(eacl.RoleOthers)

	getAllowRecord := eacl.NewRecord()
	getAllowRecord.SetOperation(eacl.OperationGet)
	getAllowRecord.SetAction(eacl.ActionAllow)
	getAllowRecord.SetTargets(toWhom)

	//getDenyRecord := eacl.NewRecord()
	//getDenyRecord.SetOperation(eacl.OperationGet)
	//getDenyRecord.SetAction(eacl.ActionDeny)
	//getDenyRecord.SetTargets(toWhom)

	putAllowRecord := eacl.NewRecord()
	putAllowRecord.SetOperation(eacl.OperationPut)
	putAllowRecord.SetAction(eacl.ActionAllow)
	putAllowRecord.SetTargets(toWhom)

	//putDenyRecord := eacl.NewRecord()
	//putDenyRecord.SetOperation(eacl.OperationPut)
	//putDenyRecord.SetAction(eacl.ActionDeny)
	//putDenyRecord.SetTargets(toWhom)

	table.SetCID(cid)
	table.AddRecord(getAllowRecord)
	table.AddRecord(putAllowRecord)
	//table.AddRecord(getDenyRecord)
	//table.AddRecord(putDenyRecord)
	//table.AddRecord(denyGETRecord)//deny must come first. Commented while debugging

	return table, nil
}

func CreateEACLTable(cnrID *cid.ID, publicKey *ecdsa.PublicKey) error {
	// Attaching extended ACL:
	// |Permit|GET|obj:Colour=Red|PublicKey:pub
	// | Deny |GET|obj:Colour=Red|OTHERS

	// extended ACL record that denies GET request to object
	// with attribute 'Colour:Red' from group 'OTHERS'
	denyRecord := eacl.NewRecord()
	denyRecord.SetAction(eacl.ActionDeny)
	denyRecord.SetOperation(eacl.OperationGet)
	denyRecord.AddObjectAttributeFilter(eacl.MatchStringEqual, "Colour", "Red")

	targetOthers := eacl.NewTarget()
	targetOthers.SetRole(eacl.RoleOthers)
	denyRecord.SetTargets(targetOthers)

	// extended ACL record that allows GET request to object
	// with attribute 'Colour:Red' from specified public key

	// 33-byte hex encoded public key from N3 wallet
	//const pubKeyStr = "03ab362a4eda62d22505ffe5a5e5422f1322317e8088afedb7c5029801e1ece806"
	//pub, err := keys.NewPublicKeyFromString(pubKeyStr)
	//if err != nil {
	//	return err
	//}

	allowRecord := eacl.NewRecord()
	allowRecord.SetAction(eacl.ActionAllow)
	allowRecord.SetOperation(eacl.OperationGet)
	allowRecord.AddObjectAttributeFilter(eacl.MatchStringEqual, "Colour", "Red")

	targetKey := eacl.NewTarget()
	eacl.SetTargetECDSAKeys(targetKey, publicKey)
	allowRecord.SetTargets(targetKey)

	// create extended ACL table with two records
	table := eacl.NewTable()
	table.SetCID(cnrID)
	//must add allow before deny
	table.AddRecord(allowRecord)
	table.AddRecord(denyRecord)
	return nil
}

//EqualRecords is used to check whether the records we attempted to create, and the records we get back, match
func EqualRecords(r1, r2 []*eacl.Record) bool {
	if len(r1) != len(r2) {
		return false
	}

	for i := 0; i < len(r1); i++ {
		d1, err := r1[i].Marshal()
		if err != nil {
			return false
		}

		d2, err := r2[i].Marshal()
		if err != nil {
			return false
		}

		if !bytes.Equal(d1, d2) {
			return false
		}
	}

	return true
}
