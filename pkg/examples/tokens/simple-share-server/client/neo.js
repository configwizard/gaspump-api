function decryptAccountData(privateKey) {
    // const myWallet = new Neon.wallet.Wallet({
    //     name: "MyWallet"
    // });
    const myAccount = new Neon.wallet.Account(
        privateKey
    );
    myAccount.publicKey = Neon.wallet.getPublicKeyFromPrivateKey(myAccount.privateKey)
    // myWallet.addAccount(myAccount);
    // We need to decrypt wallet by a password
    // await myAccount.decrypt("password") //todo accidentally encrypted key with different password. Idiot.
    console.log("privateKey", myAccount.privateKey)
    console.log("publickey", );
    console.log("WIF", myAccount.WIF);
    return myAccount
}
