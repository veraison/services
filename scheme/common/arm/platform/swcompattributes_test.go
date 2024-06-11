// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package platform

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSwCompAttributes_MakeRefAttrs(t *testing.T) {
	type fields struct {
		MeasurementType  string
		Version          string
		SignerID         []byte
		AlgID            string
		MeasurementValue []byte
	}
	type args struct {
		c      ClassAttributes
		scheme string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &SwCompAttributes{
				MeasurementType:  tt.fields.MeasurementType,
				Version:          tt.fields.Version,
				SignerID:         tt.fields.SignerID,
				AlgID:            tt.fields.AlgID,
				MeasurementValue: tt.fields.MeasurementValue,
			}
			got, err := o.MakeRefAttrs(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("SwCompAttributes.MakeRefAttrs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SwCompAttributes.MakeRefAttrs() = %v, want %v", got, tt.want)
			}
		})
	}
}
