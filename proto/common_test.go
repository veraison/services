// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_StringList_round_trip(t *testing.T) {
	vals := []string{"huey", "dewey", "louie"}

	stringList, err := NewStringList(vals)
	require.NoError(t, err)

	listValue := stringList.AsListValue()
	require.IsType(t, &structpb.ListValue{}, listValue)
	assert.Equal(t, "huey", listValue.Values[0].GetStringValue())
	assert.Equal(t, vals, stringList.AsSlice())
}
