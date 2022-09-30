This is CLI interface to the policy store. It allows typical CRUD operations on
the store as well as listing all stored policies (use -h flag for more details
of available commands).

Connection to the store is configured by "po-store" entry in a "config.yaml" in
the current directory (see the included example file). Alternatively, an
sqlite3 database file can be specified with -s/--store.


## Examples

Perform a one-time setup of a new store

    ./polcli setup

(For an SQL-backed store, this will create the required table.)

Add policy from a file under the specified ID:

    ./polcli add opa://1 path/to/policy.rego

(Note: ID must be in the form "opa://<tenant id>", where <tenant id> is the
integer ID of the tenant for whom the policy will be added. The "opa://" prefix
indicates the policy format; currently, only OPA rego policies are supported.)

Update and existing ID with a new version (or add if ID doesn't already exist):

    ./polcli add -u opa://1 path/to/newpolicy.rego

List stored policies:

    ./polcli list

(The versions listed are the latest associated with the corresponding ID.
Alternatively, -a flag can be used to list all stored versions.)

Print policy stored under the specified ID to STDOUT:

    ./polcli get opa://1

(This will print the latest version. -v flag can be used to specify an earlier
version. -o flag can be used to specify a file to write to, instead of printing
to STDOUT.)


Delete policy with the specified ID:

    ./polcli del opa://1

(This will delete all versions associated with ID from the store.)


