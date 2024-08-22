// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import "testing"

func TestNormalizeMediaType(t *testing.T) {
	type args struct {
		mt string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"already normalized",
			args{`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`},
			`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`,
			false,
		},
		{
			"mixed case",
			args{`application/EAT+CWT; eat_profile="tag:psacertified.org,2023:psa#tfm"`},
			`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`,
			false,
		},
		{
			"extra white-spaces",
			args{` application/eat+cwt  ; eat_profile="tag:psacertified.org,2023:psa#tfm"    `},
			`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`,
			false,
		},
		{
			// Absolute URIs contain ":" which is not allowed by the `token` grammar
			// therefore, they need to be quoted
			"unquoted URI parameter",
			args{`application/eat+cwt; eat_profile=tag:psacertified.org,2023:psa#tfm`},
			"",
			true,
		},
		{
			// OIDs using the dotted-decimal notation parse under the `token` grammar
			// therefore, they don't require quoting
			"unquoted OID parameter",
			args{`application/eat+cwt; eat_profile=2.999.1`},
			"application/eat+cwt; eat_profile=2.999.1",
			false,
		},
		{
			// normalisation removes the unnecessary quoting
			"quoted OID parameter",
			args{`application/eat+cwt; eat_profile="2.999.1"`},
			"application/eat+cwt; eat_profile=2.999.1",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMediaType(tt.args.mt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeMediaType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeMediaType() = %v, want %v", got, tt.want)
			}
		})
	}
}
