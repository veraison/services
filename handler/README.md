This package defines [`IEvidenceHandler`](ievidencehandler.go), 
[`IEndorsementHandler`](iendorsementhandler.go),
[`IStoreHandler`](istorehandler.go) and [`ICoservHandler`](icoservhandler.go)
[pluggable](../plugin/README.md) interfaces and associated RPC channels.
These are used to add new attestation scheme to Veraison services.
Additionally, the package defines a [couple of wrappers](plugin.go) around
`plugin.RegisterImplementation` for registering implementations of these four
interfaces.
