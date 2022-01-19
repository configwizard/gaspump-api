package wallet_test

import (
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
	assert.Nil(t, err, "error not nil")
	assert.NotNil(t, w.Accounts[0], "no account")
	assert.NotEqualf(t, "", w.Accounts[0].Address, "no address")

	creds, err := wallet.GetCredentials(path, w.Accounts[0].Address, password)
	assert.Nil(t, err, "error not nil")
	assert.NotEqual(t, nil, creds)
}
