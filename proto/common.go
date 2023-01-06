// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package proto

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

type StringList struct {
	list *structpb.ListValue
}

func ListValuetoStringList(v *structpb.ListValue) *StringList {
	return &StringList{list: v}
}

func NewStringList(vs []string) (*StringList, error) {
	var in []interface{}
	for _, v := range vs {
		in = append(in, v)
	}

	list, err := structpb.NewList(in)
	if err != nil {
		return nil, err
	}

	return &StringList{list}, nil
}

func (o StringList) AsListValue() *structpb.ListValue {
	return o.list
}

func (o StringList) AsSlice() []string {
	var out []string

	for _, v := range o.list.AsSlice() {
		out = append(out, fmt.Sprint(v))
	}

	return out
}
