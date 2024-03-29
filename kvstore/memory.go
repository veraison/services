// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"text/tabwriter"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	lk = sync.RWMutex{}
)

type Memory struct {
	Data map[string][]string

	logger *zap.SugaredLogger
}

// Init initializes the KVStore. There are no configuration options for this
// implementation.
func (o *Memory) Init(unused *viper.Viper, logger *zap.SugaredLogger) error {
	o.Data = make(map[string][]string)
	o.logger = logger

	return nil
}

func (o *Memory) Close() error {
	// no-op (the map is garbage collected)
	return nil
}

func (o *Memory) Setup() error {
	// no-op (the map is created on init, and no further setup is necessary)
	return nil
}

func (o Memory) Get(key string) ([]string, error) {
	if o.Data == nil {
		return nil, errors.New("memory store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return nil, err
	}

	lk.RLock()
	defer lk.RUnlock()

	vals, ok := o.Data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrKeyNotFound, key)
	}

	return vals, nil
}

func (o Memory) GetKeys() ([]string, error) {
	if o.Data == nil {
		return nil, errors.New("memory store uninitialized")
	}

	lk.RLock()
	defer lk.RUnlock()

	var keys []string // nolint:prealloc
	for k := range o.Data {
		keys = append(keys, k)
	}

	return keys, nil
}

func (o *Memory) Add(key string, val string) error {
	if o.Data == nil {
		return errors.New("memory store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	lk.Lock()
	defer lk.Unlock()

	data, ok := o.Data[key]
	if ok {
		// check if val is already present
		for _, d := range data {
			if d == val {
				return nil
			}
		}
		o.Data[key] = append(data, val)
	} else {
		o.Data[key] = []string{val}
	}

	return nil
}

func (o *Memory) Set(key string, val string) error {
	if o.Data == nil {
		return errors.New("memory store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	lk.Lock()
	defer lk.Unlock()

	o.Data[key] = []string{val}

	return nil
}

func (o *Memory) Del(key string) error {
	if o.Data == nil {
		return errors.New("memory store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return err
	}

	lk.Lock()
	defer lk.Unlock()

	if _, ok := o.Data[key]; !ok {
		return fmt.Errorf("%w: %q", ErrKeyNotFound, key)
	}

	delete(o.Data, key)

	return nil
}

func (o Memory) dump() string {
	var b bytes.Buffer

	w := tabwriter.NewWriter(&b, 1, 1, 1, ' ', 0)

	fmt.Fprintln(w, "Key\tVal")
	fmt.Fprintln(w, "---\t---")

	lk.RLock()
	defer lk.RUnlock()

	// stabilize output order
	sortedKeys := make([]string, 0, len(o.Data))
	for k := range o.Data {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		fmt.Fprintf(w, "%s\t%s\n", k, o.Data[k])
	}

	w.Flush()

	return b.String()
}

func (o Memory) Dump() {
	fmt.Println(o.dump())
}
