// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package compositeevidenceparser

import (
	"fmt"

	"github.com/veraison/cmw"
)

type cmwp struct{}

func (o cmwp) Parse(evidence []byte) ([]ComponentEvidence, error) {
	cmwCollection := &cmw.CMW{}

	if err := cmwCollection.Deserialize(evidence); err != nil {
		return nil, fmt.Errorf("unable to unmarshal CMW Collection: %w", err)
	}
	if err := cmwCollection.ValidateCollection(); err != nil {
		return nil, fmt.Errorf("unable to validate CMW Collection: %w", err)
	}
	ces, err := parseCMW(cmwCollection, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to parse CMW Collection: %w", err)
	}

	return ces, nil
}

func parseCMW(c *cmw.CMW, depth uint) ([]ComponentEvidence, error) {
	var ce []ComponentEvidence
	meta, err := c.GetCollectionMeta()
	if err != nil {
		return ce, fmt.Errorf("meta error at depth %d, %w", depth, err)
	}

	for i, k := range meta {
		switch k.Kind {
		case cmw.KindMonad:
			mycmw, err := c.GetCollectionItem(k.Key)
			if err != nil {
				return ce, fmt.Errorf("unable to get collectionItem %w", err)
			}
			e, err := insertEvidence(mycmw, k.Key, depth)
			if err != nil {
				return ce, fmt.Errorf("unable to insert Evidence %w", err)
			}
			ce = append(ce, *e)
		case cmw.KindCollection:
			depth++
			icmw, err := c.GetCollectionItem(k.Key)
			if err != nil {
				return ce, fmt.Errorf("unable to get collectionItem %w", err)
			}
			ces, err := parseCMW(icmw, depth)
			if err != nil {
				return ce, fmt.Errorf("unable to parse collections in a CMW: %w", err)
			}
			ce = append(ce, ces...)
		case cmw.KindUnknown:
			fallthrough
		default:
			return ce, fmt.Errorf("unknown kind, found at index %d %s", i, k.Kind.String())
		}

	}
	return ce, nil
}

func insertEvidence(c *cmw.CMW, key any, depth uint) (*ComponentEvidence, error) {
	var ev ComponentEvidence
	mt, err := c.GetMonadType()
	if err != nil {
		return nil, fmt.Errorf("invalid monad type: %w", err)
	}
	ev.mediaType = mt
	ev.depth = depth
	d, err := c.GetMonadValue()
	if err != nil {
		return nil, fmt.Errorf("invalid monad data: %w", err)
	}
	ev.data = d
	switch t := key.(type) {
	case string:
		ev.label = t
	case uint64:
		ev.label = fmt.Sprintf("##%d", t)
	default:
		return nil, fmt.Errorf("key with unknown type %T", t)
	}
	return nil, nil
}
