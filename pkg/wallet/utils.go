package wallet

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

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
