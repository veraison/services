This directory contains the boilerplate structure for a new attestation scheme
implementation. It is intended as a starting point.

To create a new attestation scheme:

1. Copy the contents of this directory into a new subdirectory under scheme,
   e.g. `cp -r scheme/example scheme/my-scheme`.
2. Update the `<TODO>` placeholders with relevant information, keeping the
   following conventions in mind:
   - `SchemeName` ought to be in all caps, with underscores separating words,
     e.g. `"MY_SCHEME"`
   - `GOPKG` should contain the scheme name in all lower case with dashes
     instead of underscores, e.g.

        GOPKG := github.com/veraison/services/scheme/my-scheme

     (note: Go automatically translates dashes to underscores in package names
     when parsing source, so your source  files should specify package as
     `package my_scheme`).
   - The name of the plugin executable should be the same name as `GOPKG` but
     prefixed with `scheme-` (this is to distinguish from CoSERV backend
     plugins) and use `.plugin` extension, e.g.

        PLUGIN := ../../bin/scheme-my-plugin.plugin

3. Update `scheme/Makefile` with a `SUBDIR` entry for your scheme.
4. Update `builtin/schemes.go` with an entry for your scheme (see existing
   entries there for an example).
