const ffi = require('ffi-cross');


var awesome = ffi.Library("./lib.dylib", {
    RetrieveNeoFSBalance: [ ffi.types.double, [ ffi.types.CString, ffi.types.CString ] ],
    Test: [ffi.types.CString, []]
});

const hello = async function() {
    let val = await awesome.Test()
    console.log(val)
}
hello()
// console.log("balance = ", awesome.RetrieveNeoFSBalance("../../pkg/examples/sample_wallets/wallet.rawContent.go", "password"));
