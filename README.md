# GasPump Api

The api holds all the backend functionality to interface with NeoFS.

If the repository is private, you will need to be a collaborator on this repository. After that you will need to generate a github access token in your personal developer settings.

Then run

```shell
go env -w GOPRIVATE=github.com/amlwwalker/gaspump-api
```

then you will need to run

```shell
git config --global url."git@github.com/amlwwalker:<YOUR ACCESS TOKEN>".insteadOf "https://github.com/amlwwalker"
```

replacing `<YOUR ACCESS TOKEN>` with your access token you retrieved earlier.

Now when building a project that uses this api, like gaspump.react you should have no issues building 
