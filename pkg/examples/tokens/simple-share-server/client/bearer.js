// const neon = require('@cityofzion/neon-js');
// import { base58_to_binary, binary_to_base58 } from 'base58-js';
// import * as grpc from '@grpc/grpc-js'
// //const grpc = require("@grpc/grpc-js");
// import { fromByteArray } from 'base64-js';
const sha512 = require('js-sha512');
const ecdsa = require('js-ecdsa')
//
// // TEMPORARILY use google auth to establish connection to NeoFS.
// var GoogleAuth = require('google-auth-library'); // from https://www.npmjs.com/package/google-auth-library
//
// import { AccountingServiceClient } from './generated/accounting/service_grpc_pb';
// const {Signature, OwnerID, Version, ContainerID} = require('./generated/refs/types_pb');
// const {BalanceRequest} = require('./generated/accounting/service_pb');
// const {RequestMetaHeader, RequestVerificationHeader} = require('./generated/session/service_pb');
//
//

function signBearerToken(payload) {
    // const client = new AccountingServiceClient('st01.testnet.fs.neo.org:8082', grpc.credentials.createSsl());
    // h := sha512.Sum512(binaryData)
    // x, y, err := ecdsa.Sign(rand.Reader, &containerOwnerPrivateKey.PrivateKey, h[:])
    // if err != nil {
    //     panic(err)
    // }
    // signatureData := elliptic.Marshal(elliptic.P256(), x, y)
    // containerOwnerPublicKeyBytes := containerOwnerPrivateKey.PublicKey().Bytes()

    const rawPrivateKey = ""
    // const publicKey = ""
    const account = new neon.wallet.Account(rawPrivateKey);
    const h = sha512.sha512.digest(payload)
    //generate p-521 ecdsa keypair
    // ecdsa.gen('521', function(err, gen){ // generates a key pair if we don't have one
    //     if(err){return console.log(err)}
    //     console.log(gen)

    //sign some data
    ecdsa.sign(account.privateKey, payload, h, "hex", function(err, sig){
        if(err){return console.log(err)}
        console.log(sig)

        //verify signature
        ecdsa.verify(account.publicKey, sig, payload, h, "hex", function(err, res){
            if(err){return console.log(err)}
            if(res){
                return console.log('ecdsa test pass')
            }
            return console.log('ecdsa test fail')
        })
    })
    // })
    /*
        const curve = u.getCurve(u.EllipticCurvePreset.SECP256R1);
    const sig = curve.sign(u.ab2hexstring(sha512.digest(msg)), account.privateKey);

    const signature = new Signature();
    const b64Key = fromByteArray(u.hexstring2ab(account.publicKey));
    signature.setKey(b64Key);
    signature.setSign(fromByteArray(u.hexstring2ab('04' + sig.r + sig.s)));
    return signature;

     */
}

modules.exports = {
    signBearerToken
}

//
// export function neofsGetBalance(ownerKey, cb, errorCb) {
//
//     const client = new AccountingServiceClient('st01.testnet.fs.neo.org:8082', grpc.credentials.createSsl());
//
//
//     const account = new wallet.Account(ownerKey);
//     const address = account.address;
//     const decoded = base58_to_binary(address);
//
//     const ownerId = new OwnerID();
//     ownerId.setValue(decoded);
//
//     const requestBody = new BalanceRequest.Body();
//     requestBody.setOwnerId(ownerId);
//
//     const metaHeader = new RequestMetaHeader();
//     metaHeader.setTtl(2);
//
//     const request = new BalanceRequest();
//     request.setBody(requestBody);
//     request.setMetaHeader(metaHeader);
//
//     const signedRequest = signRequest(request, account);
//
//     client.balance(signedRequest, {}, (err, response) => {
//       if (err) {
//         console.error(err.message);
//         errorCb(err.message);
//       } else {
//         const balance = response.getBody().getBalance().getValue() / 10 ** response.getBody().getBalance().getPrecision();
//         console.log('NeoFS contract balance is ' + balance);
//         cb(balance);
//       }
//     });
//   }
//
// function signRequest(to_sign, account) {
//
//     if (to_sign.getMetaHeader() == null) {
//         to_sign.setMetaHeader(new RequestMetaHeader());
//     }
//     const verify_origin = to_sign.getVerifyHeader();
//     const meta_header = to_sign.getMetaHeader();
//     const verify_header = new RequestVerificationHeader();
//
//     if (verify_origin == null) {
//         verify_header.setBodySignature(neofsSign(to_sign.getBody().serializeBinary(), account));
//     }
//
//     verify_header.setMetaSignature(neofsSign(meta_header.serializeBinary(), account));
//
//     if (verify_origin == null) {
//         verify_header.setOriginSignature(neofsSign(new RequestVerificationHeader().serializeBinary(), account));
//     }
//     else {
//         verify_header.setOriginSignature(neofsSign(verify_origin.serializeBinary(), account));
//     }
//     verify_header.setOrigin(verify_origin);
//     to_sign.setVerifyHeader(verify_header);
//     return to_sign;
// }
//
// function neofsSign(msg, account) {
//     const curve = u.getCurve(u.EllipticCurvePreset.SECP256R1);
//     const sig = curve.sign(u.ab2hexstring(sha512.digest(msg)), account.privateKey);
//
//     const signature = new Signature();
//     const b64Key = fromByteArray(u.hexstring2ab(account.publicKey));
//     signature.setKey(b64Key);
//     signature.setSign(fromByteArray(u.hexstring2ab('04' + sig.r + sig.s)));
//     return signature;
// }
//
// neofsGetBalance('1daa689d543606a7c033b7d9cd9ca793189935294f5920ef0a39b3ad0d00f190', function(stuff){console.log(stuff)}, function(error){console.log(error)})
