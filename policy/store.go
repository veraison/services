// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/veraison/services/kvstore"
	"go.uber.org/zap"
)

var ErrNoPolicy = errors.New("no policy found")
var ErrNoActivePolicy = errors.New("no active policy for key")

// NewStore returns a new policy store. Config options are the same as those
// used for kvstore.New().
func NewStore(v *viper.Viper, logger *zap.SugaredLogger) (*Store, error) {
	kvStore, err := kvstore.New(v, logger)
	if err != nil {
		return nil, err
	}

	return &Store{KVStore: kvStore, Logger: logger}, nil
}

type Store struct {
	KVStore kvstore.IKVStore
	Logger  *zap.SugaredLogger
}

// Setup the underyling kvstore. This is a one-time setup that only needs to be
// performed once for a deployment.
func (o *Store) Setup() error {
	return o.KVStore.Setup()
}

// Add a policy with the specified ID and rules. If a policy with that ID
// already exists, an error is returned.
func (o *Store) Add(id PolicyKey, name, typ, rules string) (*Policy, error) {
	if _, err := o.Get(id); err == nil {
		return nil, fmt.Errorf("policy with id %q already exists", id)
	}

	return o.Update(id, name, typ, rules)
}

// Update sets the provided rules as the latest version of the policy with the
// specified key. If a policy with that key does not exist, it is created.
func (o *Store) Update(key PolicyKey, name, typ, rules string) (*Policy, error) {
	newPolicy, err := NewPolicy(key, name, typ, rules)
	if err != nil {
		return newPolicy, err
	}

	return newPolicy, o.addPolicy(newPolicy)
}

// Get returns the slice of all Policies associated with the specified ID. Each
// Policy represents a different version of the same logical policy.
func (o *Store) Get(key PolicyKey) ([]*Policy, error) {
	vals, err := o.KVStore.Get(key.String())
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, fmt.Errorf("%w: %q", ErrNoPolicy, key)
		}
		return nil, err
	}

	var policies []*Policy // nolint:prealloc

	for _, v := range vals {
		var p Policy
		if err = json.Unmarshal([]byte(v), &p); err != nil {
			return nil, err
		}

		p.StoreKey = key
		policies = append(policies, &p)
	}

	return policies, nil
}

// List returns []Policy containing latest versions of all policies. All
// policies returned will have distinct IDs. In cases where multiple policies
// exist for one ID in the store, the latest version will be returned.
func (o *Store) List() ([]*Policy, error) {
	keys, err := o.GetPolicyKeys()
	if err != nil {
		return nil, err
	}

	var policies []*Policy // nolint:prealloc
	for _, key := range keys {
		policy, err := o.GetActive(key)
		if err != nil {
			if errors.Is(err, ErrNoActivePolicy) {
				continue
			}

			return nil, err
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

// ListAllVersions returns a []Policy containing every policy entry in the
// underlying store, including multiple versions associated with a single
// policy ID.
func (o *Store) ListAllVersions() ([]*Policy, error) {
	keys, err := o.GetPolicyKeys()
	if err != nil {
		return nil, err
	}

	var policies []*Policy
	for _, key := range keys {
		versions, err := o.Get(key)
		if err != nil {
			return nil, err
		}

		policies = append(policies, versions...)
	}

	return policies, nil
}

// GetPolicyKeys returns a []PolicyID of the policies currently in the store.
func (o *Store) GetPolicyKeys() ([]PolicyKey, error) {
	keys, err := o.KVStore.GetKeys()
	if err != nil {
		return nil, err
	}

	ids := make([]PolicyKey, len(keys))
	for i, k := range keys {
		key, err := PolicyKeyFromString(k)
		if err != nil {
			return nil, fmt.Errorf("bad key in store: %w", err)
		}

		ids[i] = key
	}

	return ids, nil
}

// Activate activates the policy version with the specified id for the
// specified key.
func (o *Store) Activate(key PolicyKey, id uuid.UUID) error {
	policies, err := o.Get(key)
	if err != nil {
		return err
	}

	activated := false
	for _, pol := range policies {
		if bytes.Equal(id[:], pol.UUID[:]) {
			pol.Active = true
			activated = true
		} else {
			pol.Active = false
		}
	}

	if !activated {
		return fmt.Errorf("%w with UUID %q for key %q", ErrNoPolicy, id, key.String())
	}

	if err := o.Del(key); err != nil {
		return err
	}

	for _, pol := range policies {
		if err := o.addPolicy(pol); err != nil {
			return err
		}
	}

	return nil
}

// DeactivateAll deactivates all policies associated with the key.
func (o *Store) DeactivateAll(key PolicyKey) error {
	policies, err := o.Get(key)
	if err != nil {
		return err
	}

	for _, pol := range policies {
		pol.Active = false
	}

	if err := o.Del(key); err != nil {
		return err
	}

	for _, pol := range policies {
		if err := o.addPolicy(pol); err != nil {
			return err
		}
	}

	return nil
}

// GetActive returns the current active version of the policy with the
// specified key, or an error if no such policy exists.
func (o *Store) GetActive(key PolicyKey) (*Policy, error) {
	policies, err := o.Get(key)
	if err != nil {
		return nil, err
	}

	for _, pol := range policies {
		if pol.Active {
			return pol, nil
		}
	}

	return nil, fmt.Errorf("%w %q", ErrNoActivePolicy, key.String())
}

// GetPolicy returns the policy with the specified UUID under the specified
// key.
func (o *Store) GetPolicy(key PolicyKey, id uuid.UUID) (*Policy, error) {
	policies, err := o.Get(key)
	if err != nil {
		return nil, err
	}

	for _, pol := range policies {
		if bytes.Equal(id[:], pol.UUID[:]) {
			return pol, nil
		}
	}

	return nil, fmt.Errorf("%w with UUID %q under key %q",
		ErrNoPolicy, id.String(), key.String())
}

// Del removes all policy versions associated with the specified key.
func (o *Store) Del(key PolicyKey) error {
	return o.KVStore.Del(key.String())
}

// Close the connection to the underlying kvstore.
func (o *Store) Close() error {
	return o.KVStore.Close()
}

func (o *Store) addPolicy(policy *Policy) error {
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	return o.KVStore.Add(policy.StoreKey.String(), string(policyBytes))
}
