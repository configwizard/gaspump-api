package eacl

import (
	"bytes"
	"crypto/ecdsa"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
)

//AllowOthersReadOnly from https://github.com/nspcc-dev/neofs-s3-gw/blob/fdc07b8dc15272e2aabcbd7bb8c19e435c94e392/authmate/authmate.go#L358
func AllowOthersReadOnly(cid *cid.ID) (*eacl.Table, error) {
	table := eacl.NewTable()
	record := eacl.NewRecord()
	record.SetOperation(eacl.OperationGet)
	record.SetAction(eacl.ActionAllow)
	// TODO: Change this later.
	// from := eacl.HeaderFromObject
	// matcher := eacl.MatchStringEqual
	// record.AddFilter(from eacl.FilterHeaderType, matcher eacl.Match, name string, value string)
	eacl.AddFormedTarget(record, eacl.RoleOthers)
	table.SetCID(cid)
	table.AddRecord(record)

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
