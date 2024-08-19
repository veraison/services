# KV Store

The key-values (KV) store is Veraison storage layer.  It is used for both
endorsements and trust anchors.

It is intentionally "dumb": we assume that the filtering smarts are provided by
the plugins.

The key is a string synthesised deterministically from a structured endorsement
/ trust anchor "identifier".  It is formatted according to a custom URI format
-- [see below](#uri-format)).

The value is an array of JSON strings each containing an endorsement or trust
anchor data associated with that key.  The data is opaque to the KV store and
varies depending on the attestation format.  The only invariant enforced by the
KV store is that it is valid JSON.

The `IKVStore` interface defines the required methods for storing, fetching and
deleting KV objects.  Note that (for the moment) there is no method for
patching data in place.  Interface methods for initialising and orderly
terminating the underlying DB are also exposed.

This package contains two implementations of the `IKVStore`:

1. `SQL`, supporting different SQL engines (e.g., SQLite, PostgreSQL, etc. --
   [see below](#sql-drivers)),
1. `Memory`, a thread-safe in-memory associative array intended for testing.

A `New` method can be used to create either of these from a `Config` object.

## Configuration

`kvstore` expects the following entries in configuration:


- `backend`: the name of the backend to use for the store. Currently supported
  backends: `memory`, `sql`.
- `<backend name>`: an entry with the name of a backend is used to specify the
  configuration for that backend. There may be multiple such entries for different
  backends. Only the entry matching the active backend specified by `backend`
  directive will actually be used. The contents for each entry is specific to
  the backend.

Note: in a config file, `kvstore` configuration will typically be namespaced
under the name of a particular store instance, e.g.

```yaml
ta-store:
  backend: sql
  sql:
    driver: sqlite3
```

### `memory` backend configuration

Currently, `memory` backend does not support any configuration.

### `sql` backend configuration

> **Note**: `sqlite3`, the default driver for the backend, is unsuitable for
> production or performance testing.

- `driver`: The name of the golang SQL driver. Veraison currently includes the
  following drivers: `sqlite3`, `mysql` (MySQL and MariaDB), and `pgx`
  (Postgres).
- `datasource`: This points to the database the driver will access. The format
  is driver-dependent.
  - `sqlite3`: the path to the sqlite3 database file
  - `pgx`: a URL in the form `postgresql://<user>:<passwd>@<host>:<port>/<database>`
  - `mysql`: string in the form `<user>:<password>@<protocol>(<address>)/<database>`
  (Please see the drivers' documentation for more details.)
- `tablename` (optional): the name of the table within the SQL database that will
- be used by the store. If this is not specified, it will default to
  `"kvstore"`.

## Alternative SQL drivers

To use another SQL driver ([see here](https://go.dev/wiki/SQLDrivers)) in your
code that uses `IKVStore`, the calling code needs to (anonymously) import the
supporting driver.

For example, to use PostgreSQL:
```go
import _ "github.com/lib/pq"
```
Instead, to use SQLite:
```go
import _ "github.com/mattn/go-sqlite3"
```

## SQL schemas

```sql
CREATE TABLE endorsements (
  kv_key text NOT NULL,
  kv_val text NOT NULL
);

CREATE TABLE trust_anchors (
  kv_key text NOT NULL,
  kv_val text NOT NULL
);

CREATE TABLE policies (
  kv_key text NOT NULL,
  kv_val text NOT NULL
);
```

## URI format

```abnf
scheme ":" authority path-absolute
```

where:

* `scheme` encodes the attestation format (e.g., "psa", "tcg-dice",
"tpm-enacttrust", "open-dice", "tcg-tpm", etc.)
* `authority` encodes the tenant
* `path-absolute` encodes the parts of the key, identified positionally.  Missing optional parts are encoded as empty path segments.

Attestation technology specific code (i.e., plugins) must provide their own synthesis functions.

### Examples

PSA

* Trust Anchor ID
  * `psa-iot://`TenantID.Fmt()`/`ImplID.Fmt()`/`InstID.Fmt()`
* Software ID (Model is optional)
  * `psa-iot://`TenantID.Fmt()`/`ImplID.Fmt()`/`Model.Fmt()
  * `psa-iot://`TenantID.Fmt()`/`ImplID.Fmt()`/`


EnactTrust TPM

* Trust Anchor ID
  * `tpm-enacttrust://`TenantID.Fmt()`/`NodeID.Fmt()
* Software ID
  * `tpm-enacttrust://`TenantID.Fmt()`/`NodeID.Fmt()

