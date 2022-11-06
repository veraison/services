// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

type ParsecKeyAttestation struct {
	Profile string `cbor:"265,keyasint"`
	KAT     []byte `cbor:"kat"`
	PAT     []byte `cbor:"pat"`
}

type KAT struct {
	KID      []byte `cbor:"kid" json:"kid"`
	X5C      []byte `cbor:"x5c" json:"x5c"`
	Alg      int    `cbor:"alg" json:"alg"`
	Sig      []byte `cbor:"sig" json:"sig"`
	PubArea  []byte `cbor:"pubArea" json:"pubArea"`
	CertInfo []byte `cbor:"certInfo" json:"certInfo"`
}

var (
	Nonce = []byte{
		216, 210, 247, 119, 121, 118, 109, 19, 28, 142, 6, 6, 222, 13, 177, 193,
		155, 224, 33, 181, 250, 116, 131, 8, 59, 218, 94, 243, 81, 50, 199, 2,
	}
)

func main() {
	var (
		kPath, tPath string
	)

	flag.StringVar(&kPath, "k", "key.pem", "Path of the ECDSA key used to sign the KAT.")
	flag.StringVar(&tPath, "t", "token.cbor", "Path of the CBOR encoded token")

	flag.Parse()

	kat, err := loadToken(tPath)
	if err != nil {
		log.Fatalf("loading token from %s: %v", tPath, err)
	}

	s, err := json.MarshalIndent(kat, "", "  ")
	if err != nil {
		log.Fatalf("JSON encoding gone mad: %v", err)
	}

	fmt.Println("KAT: ", string(s))

	pkey, err := loadKey(kPath)
	if err != nil {
		log.Fatalf("loading public key from %s: %v", kPath, err)
	}

	if err := verify(kat, pkey, Nonce); err != nil {
		log.Fatalf("verification failed: %v", err)
	}

	fmt.Println("ok")
}

func loadToken(tPath string) (*KAT, error) {
	data, err := os.ReadFile(tPath)
	if err != nil {
		return nil, err
	}

	var pka ParsecKeyAttestation

	if err := cbor.Unmarshal(data, &pka); err != nil {
		return nil, err
	}

	var kat KAT

	if err := cbor.Unmarshal(pka.KAT, &kat); err != nil {
		return nil, err
	}

	return &kat, nil
}

func verify(kat *KAT, nonce tpmutil.U16Bytes) error {
	attData, err := tpm2.DecodeAttestationData(kat.CertInfo)
	if err != nil {
		return fmt.Errorf("decoding certInfo: %w", err)
	}

	if attData.Magic != 0xff544347 {
		return errors.New("magic is not TPM_GENERATED_VALUE")
	}

	if attData.Type != tpm2.TagAttestCertify {
		return errors.New("type is not TPM_ST_ATTEST_CERTIFY")
	}

	if bytes.Compare(attData.ExtraData, nonce) != 0 {
		return errors.New("nonce not matched")
	}

	if attData.AttestedCertifyInfo == nil {
		return errors.New("TPMS_CERTIFY_INFO not present")
	}

	pubArea, err := tpm2.DecodePublic(kat.PubArea)
	if err != nil {
		return fmt.Errorf("decoding pubArea: %w", err)
	}

	ok, err := attData.AttestedCertifyInfo.Name.MatchesPublic(pubArea)
	if err != nil {
		return fmt.Errorf("matching pubArea against name: %w", err)
	}

	if !ok {
		return errors.New("pubArea does not match name")
	}

	// TODOTODOTODO
	// signature verification

	fmt.Printf("pubArea: %#v\n", pubArea.ECCParameters)

	eccPKey, err := tpm2PublicToGo(pubArea)
	if err != nil {
		return fmt.Errorf("converting TPM public key to go: %w", err)
	}

	fmt.Printf("%x\n", eccPKey)

	return nil
}

func tpm2PublicToGo(tpk tpm2.Public) ([]byte, error) {
	var curve elliptic.Curve

	switch tpk.ECCParameters.CurveID {
	case tpm2.CurveNISTP256:
		curve = elliptic.P256()
	case tpm2.CurveNISTP384:
		curve = elliptic.P384()
	default:
		return nil, fmt.Errorf("unsupported curve: %v", tpk.ECCParameters.CurveID)
	}

	x := tpk.ECCParameters.Point.X()
	y := tpk.ECCParameters.Point.Y()

	gpk := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	return x509.MarshalPKIXPublicKey(gpk)
}

/*
func createEKPublicECC(eccKey *ecdsa.PublicKey) (public tpm2.Public, err error) {
	public = tpm2.Public{
		Type:    tpm2.AlgECC,
		NameAlg: tpm2.AlgSHA256,
			Attributes:    defaultEKAttributes(),
			AuthPolicy:    defaultEKAuthPolicy(),
			ECCParameters: defaultECCParams(),
	}

	public = client.DefaultEKTemplateECC()
	public.ECCParameters.Point = tpm2.ECPoint{
		XRaw: eccIntToBytes(eccKey.Curve, eccKey.X),
		YRaw: eccIntToBytes(eccKey.Curve, eccKey.Y),
	}
	public.ECCParameters.CurveID, err = goCurveToCurveID(eccKey.Curve)
	return public, err
}

// ECC coordinates need to maintain a specific size based on the curve, so we
// pad the front with zeros.  This is particularly an issue for NIST-P521
// coordinates, as they are frequently missing their first byte.
func eccIntToBytes(curve elliptic.Curve, i *big.Int) []byte {
	bytes := i.Bytes()
	curveBytes := (curve.Params().BitSize + 7) / 8
	return append(make([]byte, curveBytes-len(bytes)), bytes...)
}

func goCurveToCurveID(curve elliptic.Curve) (tpm2.EllipticCurve, error) {
	switch curve.Params().Name {
	case elliptic.P224().Params().Name:
		return tpm2.CurveNISTP224, nil
	case elliptic.P256().Params().Name:
		return tpm2.CurveNISTP256, nil
	case elliptic.P384().Params().Name:
		return tpm2.CurveNISTP384, nil
	case elliptic.P521().Params().Name:
		return tpm2.CurveNISTP521, nil
	default:
		return 0, fmt.Errorf("unsupported Go curve: %v", curve.Params().Name)
	}
}
*/

/*
	buff := new(bytes.Buffer)
	endianness := binary.BigEndian

	hash := sha256.Sum256(attest)
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		fmt.Printf("ERROR: could not sign attestation data: %v\n", err)
		os.Exit(1)
	}

	sigStruct := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{HashAlg: tpm2.AlgSHA256, R: r, S: s},
	}
	sig, err := sigStruct.Encode()
	if err != nil {
		fmt.Printf("ERROR: could not encode signature: %v\n", err)
		os.Exit(1)
	}

	attestLen := uint16(len(attest))
	if err := binary.Write(buff, endianness, attestLen); err != nil {
		fmt.Printf("ERROR writing length: %v\n", err)
		os.Exit(1)
	}

	if err := binary.Write(buff, endianness, attest); err != nil {
		fmt.Printf("ERROR writing TPMS_ATTEST structure: %v\n", err)
		os.Exit(1)
	}

	if err := binary.Write(buff, endianness, sig); err != nil {
		fmt.Printf("ERROR writing signature: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, buff.Bytes(), 0600); err != nil {
		fmt.Printf("ERROR could not write %q: %v\n", outPath, err)
		os.Exit(1)
	}
}
*/
