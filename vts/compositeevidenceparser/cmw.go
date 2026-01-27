// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package compositeevidenceparser

import (
	"fmt"

	"github.com/veraison/cmw"
)

const MaxCollectionDepth = 4

type cmwParser struct{}

func (o cmwParser) Parse(evidence []byte) ([]ComponentEvidence, error) {
	cmwCollection := &cmw.CMW{}

	if err := cmwCollection.Deserialize(evidence); err != nil {
		return nil, fmt.Errorf("unable to unmarshal CMW collection: %w", err)
	}

	if cmwCollection.GetKind() != cmw.KindCollection {
		return nil, fmt.Errorf("evidence is not a CMW collection")
	}

	if err := cmwCollection.Valid(); err != nil {
		return nil, fmt.Errorf("unable to validate CMW collection: %w", err)
	}

	depth := uint(0)
	parentLabel := ""

	ces, err := parse(cmwCollection, depth, parentLabel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse CMW collection: %w", err)
	}

	return ces, nil
}

func parse(c *cmw.CMW, depth uint, parentLabel string) ([]ComponentEvidence, error) {
	var ce []ComponentEvidence

	meta, err := c.GetCollectionMeta()
	if err != nil {
		return ce, fmt.Errorf("unable to get collection meta at depth %d: %w", depth, err)
	}

	for _, k := range meta {
		switch k.Kind {
		case cmw.KindMonad:
			collectionItem, err := c.GetCollectionItem(k.Key)
			if err != nil {
				return ce, fmt.Errorf("unable to get item for key %v at depth %d: %w", k.Key, depth, err)
			}
			e, err := componentEvidenceFromMonad(collectionItem, k.Key, depth, parentLabel)
			if err != nil {
				return ce, fmt.Errorf("unable to create Component Evidence for key %v at depth %d: %w", k.Key, depth, err)
			}
			ce = append(ce, *e)
		case cmw.KindCollection:
			depth++
			if depth > MaxCollectionDepth {
				return ce, fmt.Errorf("collection depth %d: exceeded the maximum limit %d:", depth, MaxCollectionDepth)
			}
			collectionItem, err := c.GetCollectionItem(k.Key)
			if err != nil {
				return ce, fmt.Errorf("unable to get item for key %v at depth %d: %w", k.Key, depth, err)
			}
			parentLabel, err := stringify(k.Key)
			if err != nil {
				return ce, fmt.Errorf("invalid key %v at depth %d: %w", k.Key, depth, err)
			}
			ces, err := parse(collectionItem, depth, parentLabel)
			if err != nil {
				return ce, fmt.Errorf("unable to parse nested collection for key %s at depth %d: %w", parentLabel, depth, err)
			}
			ce = append(ce, ces...)
		case cmw.KindUnknown:
			fallthrough
		default:
			return ce, fmt.Errorf("unknown kind %s for key %v at depth %d", k.Kind.String(), k.Key, depth)
		}
	}

	return ce, nil
}

func componentEvidenceFromMonad(c *cmw.CMW, key any, depth uint, parentLabel string) (*ComponentEvidence, error) {
	mt, err := c.GetMonadType()
	if err != nil {
		return nil, fmt.Errorf("invalid monad type: %w", err)
	}

	d, err := c.GetMonadValue()
	if err != nil {
		return nil, fmt.Errorf("invalid monad data: %w", err)
	}

	k, err := stringify(key)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}

	return &ComponentEvidence{
		label:       k,
		data:        d,
		mediaType:   mt,
		parentLabel: parentLabel,
		depth:       depth,
	}, nil
}

func stringify(key any) (string, error) {
	switch t := key.(type) {
	case string:
		return t, nil
	case uint64:
		return fmt.Sprintf("%d", t), nil
	default:
		return "", fmt.Errorf("key with unknown type %T", t)
	}
}

func (o cmwParser) SupportedMediaTypes() []string {
	return []string{"application/cmw+cbor", "application/cmw+json"}
}
