// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
)

// Variables that contain GUIDs of AMD keys. See Section 4.1.8.1
// MSG_REPORT_REQ in the following:
// https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56421.pdf
var (
	arkGuid  = []byte{0xc0, 0xb4, 0x06, 0xa4, 0xa8, 0x03, 0x49, 0x52, 0x97, 0x43, 0x3f, 0xb6, 0x01, 0x4c, 0xd0, 0xae}
	askGuid  = []byte{0x4a, 0xb7, 0xb3, 0x79, 0xbb, 0xac, 0x4f, 0xe4, 0xa0, 0x2f, 0x05, 0xae, 0xf3, 0x27, 0xc7, 0x82}
	vcekGuid = []byte{0x63, 0xda, 0x75, 0x8d, 0xe6, 0x64, 0x45, 0x64, 0xad, 0xc5, 0xf4, 0xb9, 0x3b, 0xe8, 0xac, 0xcd}
)

// measurementByUintKey looks up comid.Measurement in a CoMID by its MKey.
//
// If no measurements are found, returns nil and no error. Otherwise,
//	returns the error encountered.
func measurementByUintKey(refVal comid.ValueTriple,
	key uint64) (*comid.Measurement, error) {
	for _, m := range refVal.Measurements.Values {
		if m.Key == nil || !m.Key.IsSet() ||
			m.Key.Type() != comid.UintType {
			continue
		}

		k, err := m.Key.GetKeyUint()
		if err != nil {
			return nil, err
		}

		if k == key {
			return &m, nil
		}
	}

	return nil, nil
}

// comidFromJson accepts a CoRIM in JSON format and returns its first CoMID
//
//	Returns error if there are more than a single CoMID, or passes on
//	error from corim routine.
func comidFromJson(buf []byte) (*comid.Comid, error) {
	extractedCorim, err := corim.UnmarshalUnsignedCorimFromJSON(buf)
	if err != nil {
		return nil, err
	}

	if len(extractedCorim.Tags) > 1 {
		return nil, errors.New("too many tags")
	}

	extractedComid, err := corim.UnmarshalComidFromCBOR(
		extractedCorim.Tags[0],
		extractedCorim.Profile,
	)

	if err != nil {
		return nil, err
	}

	return extractedComid, nil
}

// certTableEntry is an entry in the Certificate Table header in TSM report's auxblob
type certTableEntry struct {
	Guid   [16]byte
	Offset uint32
	Length uint32
}

// getKey helper to extract a particular key from the auxblob
func getKey(auxblob []byte, guid []byte) ([]byte, error) {
	for i := 0; i < len(auxblob); i += 24 {
		var entry certTableEntry
		b := auxblob[i : i+24]
		buf := bytes.NewReader(b)
		err := binary.Read(buf, binary.LittleEndian, &entry)
		if err != nil {
			return nil, err
		}

		if entry.Guid[0] == 0x0 {
			break
		}

		if bytes.Equal(guid, entry.Guid[:]) {
			return auxblob[entry.Offset : entry.Offset+entry.Length], nil
		}
	}

	return nil, fmt.Errorf("key not found: %v", guid)
}

// getARK helper function to extract ARK from auxblob
func getARK(auxblob []byte) ([]byte, error) {
	return getKey(auxblob, arkGuid)
}

// getASK helper function to extract ASK from auxblob
func getASK(auxblob []byte) ([]byte, error) {
	return getKey(auxblob, askGuid)
}

// getVCEK helper function to extract VCEK from auxblob
func getVCEK(auxblob []byte) ([]byte, error) {
	return getKey(auxblob, vcekGuid)
}
