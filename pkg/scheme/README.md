This package defines [`ISchemeHandler`](ischemehandler.go) and
[`IEndorsementStore`](iendorsementstore.go) interfaces and associated RPC
channels. These are used to add new attestation scheme and store backends
to Veraison services. Additionally, the package defines a
[couple of wrappers](plugin.go) around
`plugin.RegisterImplementation` for registering implementations of these four
interfaces.


## Notes on Endorsement Store interface

The interface has two classes of methods:

### Get endorsements

1. `GetKeyTriples` and `GetValueTriples`: query for comid key and value
triples using comid environment map as input. These methods also take a label
as input, which is used to identify the attestation scheme. These methods
must return `ErrUnsupported` if the store backend does not support the method.
This could be because of either the store backend not implementing the method
or not supporting the attestation scheme. If the store backend supports the
method and the scheme, but no matches could be found, these methods must return
`ErrNotFound`.

2. `ExecuteCoservQuery`: obtain CoSERV results as output using the base64url
encoded CoSERV query as input. This method also takes a profile as input, which
is used to identify the CoSERV profile. This method must return `ErrUnsupported`
if the store backend does not support the method or if the CoSERV profile is
not supported by the store backend. If the store backend supports the method and
the CoSERV profile, but no match could be found for the requested query, this
method must return `ErrNotFound`. This method should never return a CoSERV
containing empty results (empty AKQ or RVQ).


### Submit endorsements

The method `AddCorimBytes` takes the CoRIM bytes as input along with the
scheme. The implementation must return `ErrUnsupported` if it does not support
the operation or the attestation scheme.
