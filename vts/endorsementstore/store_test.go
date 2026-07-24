// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsementstore

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	mockstore "github.com/veraison/services/vts/endorsementstore/mocks"
)

var (
	errUnsupp     = fmt.Errorf("error: %w", handler.ErrUnsupported)
	errNotFound   = fmt.Errorf("error: %w", handler.ErrNotFound)
	errUnexpected = fmt.Errorf("unexpected error!!")
)

func TestGetKeyTriples(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeNotFound := notFoundStore(t, ctrl)
	storeUnsupported := unsupportedStore(t, ctrl)
	storeUnexpected := unexpStore(t, ctrl)
	storeValid := validStore(t, ctrl)

	store0 := VtsEndorsementStore{
		log.Named("store0"),
		[]handler.IEndorsementStorePlugin{},
	}

	store1 := VtsEndorsementStore{
		log.Named("store1"),
		nil,
	}
	store1.addStore(storeNotFound)
	store1.addStore(storeUnsupported)
	store1.addStore(storeNotFound)

	store2 := VtsEndorsementStore{
		log.Named("store2"),
		[]handler.IEndorsementStorePlugin{},
	}
	store2.addStore(storeUnsupported)
	store2.addStore(storeUnsupported)

	store3 := VtsEndorsementStore{
		log.Named("store3"),
		[]handler.IEndorsementStorePlugin{},
	}
	store3.addStore(storeNotFound)
	store3.addStore(storeUnsupported)
	store3.addStore(storeUnexpected)

	store4 := VtsEndorsementStore{
		log.Named("store4"),
		[]handler.IEndorsementStorePlugin{},
	}
	store4.addStore(storeNotFound)
	store4.addStore(storeUnsupported)
	store4.addStore(storeValid)

	_, e0 := store0.GetKeyTriples(nil, "", false)
	_, e1 := store1.GetKeyTriples(nil, "", false)
	_, e2 := store2.GetKeyTriples(nil, "", false)
	_, e3 := store3.GetKeyTriples(nil, "", false)
	_, e4 := store4.GetKeyTriples(nil, "", false)

	assert.ErrorIs(t, e0, ErrNoStores)
	assert.ErrorIs(t, e1, handler.ErrNotFound, "expected ENOTFOUND")
	assert.ErrorIs(t, e2, handler.ErrUnsupported, "expected EUNSUPPORTED")
	assert.ErrorContains(t, e3, errUnexpected.Error())
	assert.NoError(t, e4)
}

func TestGetValueTriples(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeNotFound := notFoundStore(t, ctrl)
	storeUnsupported := unsupportedStore(t, ctrl)
	storeUnexpected := unexpStore(t, ctrl)
	storeValid := validStore(t, ctrl)

	store0 := VtsEndorsementStore{
		log.Named("store0"),
		[]handler.IEndorsementStorePlugin{},
	}

	store1 := VtsEndorsementStore{
		log.Named("store1"),
		[]handler.IEndorsementStorePlugin{},
	}
	store1.addStore(storeNotFound)
	store1.addStore(storeUnsupported)
	store1.addStore(storeNotFound)

	store2 := VtsEndorsementStore{
		log.Named("store2"),
		[]handler.IEndorsementStorePlugin{},
	}
	store2.addStore(storeUnsupported)
	store2.addStore(storeUnsupported)

	store3 := VtsEndorsementStore{
		log.Named("store3"),
		[]handler.IEndorsementStorePlugin{},
	}
	store3.addStore(storeNotFound)
	store3.addStore(storeUnsupported)
	store3.addStore(storeUnexpected)

	store4 := VtsEndorsementStore{
		log.Named("store4"),
		[]handler.IEndorsementStorePlugin{},
	}
	store4.addStore(storeNotFound)
	store4.addStore(storeUnsupported)
	store4.addStore(storeValid)

	_, e0 := store0.GetValueTriples(nil, "", false)
	_, e1 := store1.GetValueTriples(nil, "", false)
	_, e2 := store2.GetValueTriples(nil, "", false)
	_, e3 := store3.GetValueTriples(nil, "", false)
	_, e4 := store4.GetValueTriples(nil, "", false)

	assert.ErrorIs(t, e0, ErrNoStores)
	assert.ErrorIs(t, e1, handler.ErrNotFound, "expected ENOTFOUND")
	assert.ErrorIs(t, e2, handler.ErrUnsupported, "expected EUNSUPPORTED")
	assert.ErrorContains(t, e3, errUnexpected.Error())
	assert.NoError(t, e4)
}

func TestExecuteCoservQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeNotFound := notFoundStore(t, ctrl)
	storeUnsupported := unsupportedStore(t, ctrl)
	storeUnexpected := unexpStore(t, ctrl)
	storeValid := validStore(t, ctrl)

	store0 := VtsEndorsementStore{
		log.Named("store0"),
		[]handler.IEndorsementStorePlugin{},
	}

	store1 := VtsEndorsementStore{
		log.Named("store1"),
		[]handler.IEndorsementStorePlugin{},
	}
	store1.addStore(storeNotFound)
	store1.addStore(storeUnsupported)
	store1.addStore(storeNotFound)

	store2 := VtsEndorsementStore{
		log.Named("store2"),
		[]handler.IEndorsementStorePlugin{},
	}
	store2.addStore(storeUnsupported)
	store2.addStore(storeUnsupported)

	store3 := VtsEndorsementStore{
		log.Named("store3"),
		[]handler.IEndorsementStorePlugin{},
	}
	store3.addStore(storeNotFound)
	store3.addStore(storeUnsupported)
	store3.addStore(storeUnexpected)

	store4 := VtsEndorsementStore{
		log.Named("store4"),
		[]handler.IEndorsementStorePlugin{},
	}
	store4.addStore(storeNotFound)
	store4.addStore(storeUnsupported)
	store4.addStore(storeValid)

	_, e0 := store0.ExecuteCoservQuery("", "")
	_, e1 := store1.ExecuteCoservQuery("", "")
	_, e2 := store2.ExecuteCoservQuery("", "")
	_, e3 := store3.ExecuteCoservQuery("", "")
	_, e4 := store4.ExecuteCoservQuery("", "")

	assert.ErrorIs(t, e0, ErrNoStores)
	assert.ErrorIs(t, e1, handler.ErrNotFound, "expected ENOTFOUND")
	assert.ErrorIs(t, e2, handler.ErrUnsupported, "expected EUNSUPPORTED")
	assert.ErrorContains(t, e3, errUnexpected.Error())
	assert.NoError(t, e4)
}

func TestAddCorimBytes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeUnexpected := unexpStore(t, ctrl)
	storeValid := validStore(t, ctrl)

	store0 := VtsEndorsementStore{
		log.Named("store0"),
		[]handler.IEndorsementStorePlugin{},
	}

	store1 := VtsEndorsementStore{
		log.Named("store1"),
		[]handler.IEndorsementStorePlugin{},
	}
	store1.addStore(storeUnexpected)
	store1.addStore(storeValid)

	store2 := VtsEndorsementStore{
		log.Named("store2"),
		[]handler.IEndorsementStorePlugin{},
	}
	store2.addStore(storeValid)
	store2.addStore(storeUnexpected)

	e0 := store0.AddCorimBytes(make([]byte, 0), "", false)
	e1 := store1.AddCorimBytes(make([]byte, 0), "", false)
	e2 := store2.AddCorimBytes(make([]byte, 0), "", false)

	assert.ErrorIs(t, e0, ErrNoStores)
	assert.Error(t, e1)
	assert.NoError(t, e2)
}

func TestComputeStoreErr(t *testing.T) {
	m := make(map[error]bool, 2)
	m[handler.ErrUnsupported] = true
	assert.ErrorIs(t, computeStoreErr(m), handler.ErrUnsupported)

	m[handler.ErrNotFound] = true
	assert.ErrorIs(t, computeStoreErr(m), handler.ErrNotFound)
}

func TestLogOrErr(t *testing.T) {
	var res error

	res = logOrErr(handler.ErrNotFound, log.Named("temp"), "", &map[error]bool{})
	assert.NoError(t, res)

	res = logOrErr(handler.ErrUnsupported, log.Named("temp"), "", &map[error]bool{})
	assert.NoError(t, res)

	res = logOrErr(errUnexpected, log.Named("temp"), "", &map[error]bool{})
	assert.ErrorIs(t, res, errUnexpected)
}

func TestStoreList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store0 := VtsEndorsementStore{
		log.Named("store0"),
		[]handler.IEndorsementStorePlugin{},
	}

	store1 := VtsEndorsementStore{
		log.Named("store0"),
		[]handler.IEndorsementStorePlugin{},
	}
	store1.addStore(validStore(t, ctrl))

	_, e0 := store0.storeList()
	_, e1 := store1.storeList()

	assert.ErrorIs(t, e0, ErrNoStores)
	assert.NoError(t, e1)
}

func notFoundStore(t *testing.T, ctrl *gomock.Controller) *mockstore.MockIEndorsementStorePlugin {
	return errStore(t, ctrl, errNotFound)
}

func unsupportedStore(t *testing.T, ctrl *gomock.Controller) *mockstore.MockIEndorsementStorePlugin {
	return errStore(t, ctrl, errUnsupp)
}

func unexpStore(t *testing.T, ctrl *gomock.Controller) *mockstore.MockIEndorsementStorePlugin {
	return errStore(t, ctrl, errUnexpected)
}

func errStore(t *testing.T, ctrl *gomock.Controller, err error) *mockstore.MockIEndorsementStorePlugin {
	storeUnexpected := mockstore.NewMockIEndorsementStorePlugin(ctrl)
	storeUnexpected.EXPECT().
		GetKeyTriples(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil, err)
	storeUnexpected.EXPECT().
		GetValueTriples(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil, err)
	storeUnexpected.EXPECT().
		ExecuteCoservQuery(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil, err)
	storeUnexpected.EXPECT().
		AddCorimBytes(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(err)
	storeUnexpected.EXPECT().
		GetName().
		AnyTimes().
		Return("error-store")

	return storeUnexpected
}

func validStore(t *testing.T, ctrl *gomock.Controller) *mockstore.MockIEndorsementStorePlugin {
	storeValid := mockstore.NewMockIEndorsementStorePlugin(ctrl)
	storeValid.EXPECT().
		GetKeyTriples(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return([]*comid.KeyTriple{&comid.KeyTriple{}}, nil)
	storeValid.EXPECT().
		GetValueTriples(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return([]*comid.ValueTriple{&comid.ValueTriple{}}, nil)
	storeValid.EXPECT().
		ExecuteCoservQuery(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&coserv.Coserv{}, nil)
	storeValid.EXPECT().
		AddCorimBytes(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)
	storeValid.EXPECT().
		GetName().
		AnyTimes().
		Return("valid-store")

	return storeValid
}
