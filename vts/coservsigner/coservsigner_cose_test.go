package coservsigner

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKey = []byte(`{
	"kty":"EC",
	"crv":"P-256",
	"x":"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4",
	"y":"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM",
	"d":"870MB6gfuTJ4HtUnUvYMyJpr5eUZNP4Bk43bVdj3eAE"
}`)

func makeFS(t *testing.T) afero.Fs {
	fs := afero.NewMemMapFs()

	err := fs.MkdirAll("testdata", 0700)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "testdata/ec256key.jwk", testKey, 0600)
	require.NoError(t, err)

	return fs
}

func TestCOSESigner_Init_OK(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/ec256key.jwk",
		Alg: "ES256",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.NoError(t, err)
}
