// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/config"
)

var FallbackMaxExpiry = 5 * time.Minute

type Context struct {
	Signer            ISigner
	FallbackAuthority *comid.CryptoKey
	MaxExpiry         time.Duration
}

func NewCoservContextFromViper(v *viper.Viper) (*Context, error) {
	expiry, err := getMaxExpiry(v)
	if err != nil {
		return nil, fmt.Errorf("max expiry: %w", err)
	}

	subs, err := config.GetSubs(v, "signer")
	if err != nil {
		return nil, fmt.Errorf("accessing sub configs: %w", err)
	}

	signer, err := NewSigner(subs["signer"], afero.NewOsFs())
	if err != nil {
		return nil, fmt.Errorf("initializing signer: %w", err)
	}

	authority, err := signer.GetAuthority()
	if err != nil {
		return nil, fmt.Errorf("fallback authority: %w", err)
	}

	return &Context{
		Signer:            signer,
		MaxExpiry:         expiry,
		FallbackAuthority: authority,
	}, nil

}

func (o *Context) Close() error {
	return o.Signer.Close()
}

func getMaxExpiry(v *viper.Viper) (time.Duration, error) {
	if !v.IsSet("max-expiry") {
		return FallbackMaxExpiry, nil
	}

	return parseDuration(v.GetString("max-expiry"))
}

func parseDuration(text string) (time.Duration, error) {
	text = strings.ToLower(text)

	re, err := regexp.Compile(`\s*(\d+)\s*(second|sec|s|minute|min|m|hour|h|day|d)s?\s*`)
	if err != nil {
		panic(err)
	}

	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("could not match time specifiers in %q", text)
	}

	// Ensure that match indexes cover the entire input; i.e. there is no
	// unrecognized text left after matching.
	indexes := re.FindAllStringIndex(text, -1)
	pos := 0
	for _, idx := range indexes {
		if idx[0] != pos {
			return 0, fmt.Errorf("parsing %q: unrecognized text between indexes %d and %d",
				text, pos, idx[0])
		}

		pos = idx[1]
	}
	if pos != len(text) {
		return 0, fmt.Errorf("parsing %q: unrecognized trailing text", text)
	}

	var ret time.Duration
	for _, match := range matches {
		val, err := strconv.Atoi(match[1])
		if err != nil {
			// cannot get here because we matched \d+
			panic(err)
		}
		duration := time.Duration(val)

		switch match[2] {
		case "s", "sec", "second":
			ret += duration * time.Second
		case "m", "min", "minute":
			ret += duration * time.Minute
		case "h", "hour":
			ret += duration * time.Hour
		case "d", "day":
			ret += duration * 24 * time.Hour
		default:
			// cannot get here because cases exhaust possible matches
			panic(fmt.Errorf("unexpected time preiod specifier %q", match[2]))
		}
	}

	return ret, nil
}
