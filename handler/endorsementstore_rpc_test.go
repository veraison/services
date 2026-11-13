// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"net"
	"net/rpc"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
)

// coservStore is a store plugin that only implements ExecuteCoservQuery.
type coservStore struct {
	IEndorsementStorePlugin

	res *coserv.Coserv
	err error
}

func (o coservStore) ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error) {
	return o.res, o.err
}

// newStoreRPCClient serves impl over net/rpc, as the go-plugin backend does,
// and returns a client connected to it.
func newStoreRPCClient(t *testing.T, impl IEndorsementStorePlugin) *StoreRPCClient {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("Plugin", getStoreServer(impl)))

	clientConn, serverConn := net.Pipe()
	go server.ServeConn(serverConn)

	client := rpc.NewClient(clientConn)
	t.Cleanup(func() { _ = client.Close() })

	return getStoreClient(client).(*StoreRPCClient)
}

func emptyCoserv(t *testing.T) *coserv.Coserv {
	t.Helper()

	artifactType := coserv.ArtifactTypeReferenceValues
	resultType := coserv.ResultTypeCollectedArtifacts
	c, err := coserv.NewCoserv("http://example.com", coserv.Query{
		ArtifactType: &artifactType,
		EnvironmentSelector: coserv.NewEnvironmentSelector().
			AddClass(coserv.StatefulClass{
				Class: comid.NewClassBytes([]byte{0x01, 0x02, 0x03, 0x04}),
			}),
		ResultType: &resultType,
	})
	require.NoError(t, err)
	require.NoError(t, c.AddResults(*coserv.NewResultSet().SetExpiry(time.Now().Add(time.Hour))))

	return c
}

func Test_StoreRPC_ExecuteCoservQuery(t *testing.T) {
	errUnexpected := errors.New("unexpected error")

	type TestVector struct {
		Title      string
		Res        *coserv.Coserv
		Err        error
		ExpectRes  bool
		ExpectErr  error
		ExpectText string
	}
	tvs := []TestVector{
		{
			Title:     "result set",
			Res:       emptyCoserv(t),
			ExpectRes: true,
		},
		{
			Title:     "empty result set with not found",
			Res:       emptyCoserv(t),
			Err:       ErrNotFound,
			ExpectRes: true,
			ExpectErr: ErrNotFound,
		},
		{
			Title:     "not found without a result set",
			Err:       ErrNotFound,
			ExpectErr: ErrNotFound,
		},
		{
			Title:      "unexpected error",
			Res:        emptyCoserv(t),
			Err:        errUnexpected,
			ExpectText: errUnexpected.Error(),
		},
	}

	for _, tv := range tvs {
		t.Run(tv.Title, func(t *testing.T) {
			client := newStoreRPCClient(t, coservStore{res: tv.Res, err: tv.Err})

			res, err := client.ExecuteCoservQuery("", "")

			switch {
			case tv.ExpectErr != nil:
				assert.ErrorIs(t, err, tv.ExpectErr)
			case tv.ExpectText != "":
				assert.EqualError(t, err, tv.ExpectText)
			default:
				assert.NoError(t, err)
			}
			if tv.ExpectRes {
				require.NotNil(t, res)
				assert.NotNil(t, res.Results)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}
