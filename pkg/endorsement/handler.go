// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsement

import (
	"errors"
	"fmt"
	"slices"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/plugin"
)

type Capability int8

const (
	GetTriples Capability = iota
	ExecuteCoservQuery
	UpdateStore
)

type IHandler interface {
	plugin.IPluggable

	// Can returns true if this implementation supports all specified capabilities
	Can(...Capability) bool

	ITripleGetter
	ICoservQuerier
	IStoreUpdater
}

type ITripleGetter interface {
	GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error)
	GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error)
}

type ICoservQuerier interface {
	ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error)
}

type IStoreUpdater interface {
	AddCorimBytes(data []byte, scheme string, activate bool) error
	SetEndorsementsState(label string, request []byte, setActive bool) error
}

type INamed interface {
	GetName() string
}

func MustWrapHandler(toWrap INamed, capabilities ...Capability) IHandler {
	ret, err := WrapHandler(toWrap, capabilities...)
	if err != nil {
		panic(err)
	}

	return ret
}

func WrapHandler(toWrap INamed, capabilities ...Capability) (IHandler, error) {
	if len(capabilities) == 0 {
		return nil, errors.New("no capabilities specified")
	}

	for _, capability := range capabilities {
		switch capability {
		case GetTriples:
			if _, ok := toWrap.(ITripleGetter); !ok {
				return nil, fmt.Errorf("%s does not implement ITripleGetter", toWrap.GetName())
			}
		case ExecuteCoservQuery:
			if _, ok := toWrap.(ICoservQuerier); !ok {
				return nil, fmt.Errorf("%s does not implement ICoservQuerier", toWrap.GetName())
			}
		case UpdateStore:
			if _, ok := toWrap.(IStoreUpdater); !ok {
				return nil, fmt.Errorf("%s does not implement IStoreUpdater", toWrap.GetName())
			}
		default:
			return nil, fmt.Errorf("unexpected capability %d", capability)
		}
	}

	return &HandlerWrapper{toWrap, capabilities}, nil
}

type HandlerWrapper struct {
	handler      INamed
	capabilities []Capability
}

func (o *HandlerWrapper) Can(capabilities ...Capability) bool {
	for _, capability := range capabilities {
		if !slices.Contains(o.capabilities, capability) {
			return false
		}
	}

	return true
}

func (o *HandlerWrapper) Init(params *plugin.Parameters) error {
	impl, ok := o.handler.(interface {
		Init(params *plugin.Parameters) error
	})
	if ok {
		return impl.Init(params)
	}

	return nil
}

func (o *HandlerWrapper) GetName() string {
	return o.handler.GetName()
}

func (o *HandlerWrapper) Fini() error {
	impl, ok := o.handler.(interface{ Fini() error })
	if ok {
		return impl.Fini()
	}

	return nil
}

func (o *HandlerWrapper) GetKeyTriples(
	env *comid.Environment,
	scheme string,
	exact bool,
) ([]*comid.KeyTriple, error) {
	if !o.Can(GetTriples) {
		return nil, fault.ErrNotSupported
	}

	return o.handler.(ITripleGetter).GetKeyTriples(env, scheme, exact)
}

func (o *HandlerWrapper) GetValueTriples(
	env *comid.Environment,
	scheme string,
	exact bool,
) ([]*comid.ValueTriple, error) {
	if !o.Can(GetTriples) {
		return nil, fault.ErrNotSupported
	}

	return o.handler.(ITripleGetter).GetValueTriples(env, scheme, exact)
}

func (o *HandlerWrapper) ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error) {
	if !o.Can(ExecuteCoservQuery) {
		return nil, fault.ErrNotSupported
	}

	return o.handler.(ICoservQuerier).ExecuteCoservQuery(profile, query)
}

func (o *HandlerWrapper) AddCorimBytes(data []byte, scheme string, activate bool) error {
	if !o.Can(UpdateStore) {
		return fault.ErrNotSupported
	}

	return o.handler.(IStoreUpdater).AddCorimBytes(data, scheme, activate)
}

func (o *HandlerWrapper) SetEndorsementsState(label string, request []byte, setActive bool) error {
	if !o.Can(UpdateStore) {
		return fault.ErrNotSupported
	}

	return o.handler.(IStoreUpdater).SetEndorsementsState(label, request, setActive)
}
