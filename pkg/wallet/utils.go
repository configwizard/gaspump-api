package wallet

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/encoding/base58"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

const (
	// NEO2Prefix is the first byte of address for NEO2.
	NEO2Prefix byte = 0x17
	// NEO3Prefix is the first byte of address for NEO3.
	NEO3Prefix byte = 0x35
)

// Prefix is the byte used to prepend to addresses when encoding them, it can
// be changed and defaults to 53 (0x35), the standard NEO prefix.
var Prefix = NEO3Prefix

func OwnerIDFromPrivateKey(key *ecdsa.PrivateKey) (*owner.ID, error) {
	w, err := owner.NEO3WalletFromPublicKey(&key.PublicKey)
	if err != nil {
		return &owner.ID{}, err
	}

	return owner.NewIDFromNeo3Wallet(w), nil
}

func PrettyPrint(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}


func gasToken(cli client.Client) (util.Uint160, error){
	gasToken, err := cli.GetNativeContractHash(nativenames.Gas)
	return gasToken, err
}

//converting addresses
//https://github.com/nspcc-dev/neo-go/blob/613a23cc3f6c303882a81b61f3baec39b7e84597/pkg/encoding/address/address.go

// Uint160ToString returns the "NEO address" from the given Uint160.
func Uint160ToString(u util.Uint160) string {
	// Dont forget to prepend the Address version 0x17 (23) A
	b := append([]byte{Prefix}, u.BytesBE()...)
	return base58.CheckEncode(b)
}

// StringToUint160 attempts to decode the given NEO address string
// into an Uint160.
func StringToUint160(s string) (u util.Uint160, err error) {
	b, err := base58.CheckDecode(s)
	if err != nil {
		return u, err
	}
	if b[0] != Prefix {
		return u, errors.New("wrong address prefix")
	}
	return util.Uint160DecodeBytesBE(b[1:21])
}
