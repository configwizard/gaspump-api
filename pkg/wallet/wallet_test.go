package wallet_test

import (
	"fmt"
	"github.com/amlwwalker/gaspump-api/pkg/wallet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWalletGenerateNew(t *testing.T) {
	w, err := wallet.GenerateNewWallet("/tmp/wallet.json")
	assert.Nil(t, err, "error not nil")
	assert.NotNil(t, w.Accounts[0], "no account")
	assert.NotEqualf(t, "", w.Accounts[0].Address, "no address")
}

func TestWalletSecureGenerateNew(t *testing.T) {
	path := "/tmp/wallet.json"
	password := "password"
	w, err := wallet.GenerateNewSecureWallet(path, "", password)
	fmt.Print(wallet.PrettyPrint(w))
	assert.Nil(t, err, "error not nil")
	assert.NotNil(t, w.Accounts[0], "no account")
	assert.NotEqualf(t, "", w.Accounts[0].Address, "no address")

	creds, err := wallet.GetCredentialsFromPath(path, w.Accounts[0].Address, password)
	assert.Nil(t, err, "error not nil")
	assert.NotEqual(t, nil, creds)
	creds, err = wallet.GetCredentialsFromWallet("", "password", w)
	assert.Nil(t, err, "error not nil")
	assert.NotEqual(t, nil, creds)
}
