# GasPump Api

The api holds all the backend functionality to interface with NeoFS.


**You can `go get` this library with** `go get github.com/configwizard/gaspump-api`


### If the repository is private

you will need to be a collaborator on this repository. After that you will need to generate a github access token in your personal developer settings.

Then run

```shell
go env -w GOPRIVATE=github.com/configwizard/gaspump-api
```

then you will need to run

```shell
git config --global url."git@github.com/<username>:<YOUR ACCESS TOKEN>".insteadOf "https://github.com/<username>"
```

replacing `<YOUR ACCESS TOKEN>` with your access token you retrieved earlier.

Now when building a project that uses this api, like gaspump.react you should have no issues building 

## Examples

Run any of the examples, they don't rely on anything except the api and what is contained in each file.

Using the documentation at [neo-doc](https://amlwwalker.github.io/neo-docs/introduction/) you should get on just fine.
