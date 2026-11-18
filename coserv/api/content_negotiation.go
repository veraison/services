// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"mime"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// acceptHeader is a parsed media type, either an entry of the supported list
// or an entry of an Accept header.
type acceptHeader struct {
	MediaType string  // e.g. application/coserv+cbor
	Profile   string  // empty if absent
	Q         float64 // Accept q-value, 1 by default
}

func parseAcceptHeader(s string) (acceptHeader, error) {
	t, params, err := mime.ParseMediaType(strings.TrimSpace(s))
	if err != nil {
		return acceptHeader{}, err
	}

	accept := acceptHeader{MediaType: t, Profile: params["profile"], Q: 1.0}

	if qs, ok := params["q"]; ok {
		q, err := strconv.ParseFloat(qs, 64)
		if err != nil || q < 0 || q > 1 {
			return acceptHeader{}, fmt.Errorf("bad q-value %q", qs)
		}
		accept.Q = q
	}

	return accept, nil
}

// splitAccept splits an Accept header on commas that are not inside double
// quotes. This is needed because profile values may themselves contain commas,
// e.g. profile="tag:arm.com,2025:cca_platform#1.0.0".
func splitAccept(h string) []string {
	var out []string

	start, inQuote := 0, false

	for i := 0; i < len(h); i++ {
		switch h[i] {
		case '"':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				out = append(out, h[start:i])
				start = i + 1
			}
		}
	}

	return append(out, h[start:])
}

// negotiateMediaType returns the media type (carrying the query's profile) that
// the response should use, or an error that must be reported as 406 Not
// Acceptable. supported is the list of media types advertised by the discovery
// document; queryProfile is the profile found in the CoSERV query.
func negotiateMediaType(accept string, supported []string, queryProfile string) (string, error) {
	var (
		candidates []acceptHeader // supported entries for the query's profile
		profiles   []string       // every supported profile
	)

	for _, s := range supported {
		supportedAccept, err := parseAcceptHeader(s)
		if err != nil {
			continue
		}

		profiles = append(profiles, supportedAccept.Profile)

		if supportedAccept.Profile == queryProfile {
			candidates = append(candidates, supportedAccept)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("unsupported CoSERV profile %q; supported profiles: %s",
			queryProfile, supportedProfiles(profiles))
	}

	// No Accept header: any representation is acceptable, use the first one.
	if strings.TrimSpace(accept) == "" {
		return coservMediaType(candidates[0].MediaType, queryProfile)
	}

	var wanted []acceptHeader

	for _, part := range splitAccept(accept) {
		spec, err := parseAcceptHeader(part)
		if err != nil || spec.Q <= 0 {
			continue
		}

		wanted = append(wanted, spec)
	}

	// Highest q-value first; header order breaks ties.
	sort.SliceStable(wanted, func(i, j int) bool { return wanted[i].Q > wanted[j].Q })

	for _, w := range wanted {
		for _, cand := range candidates {
			typeOK := w.MediaType == cand.MediaType || w.MediaType == "*/*" || w.MediaType == "application/*"
			profileOK := w.Profile == "" || w.Profile == queryProfile

			if typeOK && profileOK {
				return coservMediaType(cand.MediaType, queryProfile)
			}
		}
	}

	return "", fmt.Errorf("no acceptable representation for profile %q in Accept header; supported profiles: %s",
		queryProfile, supportedProfiles(profiles))
}

func supportedProfiles(profiles []string) string {
	sorted := slices.Clone(profiles)
	slices.Sort(sorted)

	return strings.Join(slices.Compact(sorted), ", ")
}

// coservMediaType formats a media type carrying the given CoSERV profile
// parameter, quoting and escaping the value as required by RFC 9110.
func coservMediaType(mediaType, profile string) (string, error) {
	mt := mime.FormatMediaType(mediaType, map[string]string{"profile": profile})
	if mt == "" {
		return "", fmt.Errorf("cannot format media type %q with profile %q", mediaType, profile)
	}

	return mt, nil
}
