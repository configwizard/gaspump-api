package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

func OwnerIDFromPrivateKey(key *ecdsa.PrivateKey) *owner.ID {
	w, err := owner.NEO3WalletFromPublicKey(&key.PublicKey)
	if err != nil {
		panic(fmt.Errorf("invalid private key"))
	}

	return owner.NewIDFromNeo3Wallet(w)
}
