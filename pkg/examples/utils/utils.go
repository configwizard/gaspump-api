package utils

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/cli/input"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

//
//import (
//	"crypto/ecdsa"
//	"errors"
//	"fmt"
//	"github.com/nspcc-dev/neo-go/cli/flags"
//	"github.com/nspcc-dev/neo-go/cli/input"
//	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
//	"github.com/nspcc-dev/neo-go/pkg/util"
//	"github.com/nspcc-dev/neo-go/pkg/wallet"
//	"github.com/nspcc-dev/neofs-sdk-go/owner"
//)
//
//func GetCredentials(path, address string) (*ecdsa.PrivateKey, error) {
//	w, err := wallet.NewWalletFromFile(path)
//	if err != nil {
//		return nil, fmt.Errorf("can't read the wallet: %walletPath", err)
//	}
//
//	return getKeyFromWallet(w, address)
//}

// getKeyFromWallet fetches private key from neo-go wallet structure
func getKeyFromWallet(w *wallet.Wallet, addrStr string) (*ecdsa.PrivateKey, error) {
	var (
		addr util.Uint160
		err  error
	)

	if addrStr == "" {
		addr = w.GetChangeAddress()
	} else {
		addr, err = flags.ParseAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid wallet address %s: %w", addrStr, err)
		}
	}

	acc := w.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("invalid wallet address %s: %w", addrStr, err)
	}

	pass, err := input.ReadPassword("Enter password > ")
	if err != nil {
		return nil, errors.New("invalid password")
	}

	if err := acc.Decrypt(pass, keys.NEP2ScryptParams()); err != nil {
		return nil, errors.New("invalid password")

	}

	return &acc.PrivateKey().PrivateKey, nil
}
//
//func OwnerIDFromPrivateKey(key *ecdsa.PrivateKey) *owner.ID {
//	w, err := owner.NEO3WalletFromPublicKey(&key.PublicKey)
//	if err != nil {
//		panic(fmt.Errorf("invalid private key"))
//	}
//
//	return owner.NewIDFromNeo3Wallet(w)
//}
//func GenerateNewSecureWallet(path, name, password string) (*wallet.Wallet, error) {
//	w, err := wallet.NewWallet(path)
//	w.CreateAccount(name, password)
//	return w, err
//}
