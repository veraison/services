// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/cobra"
	"github.com/veraison/ccatoken"
	"github.com/veraison/corim/comid"
	"github.com/veraison/eat"
	"github.com/veraison/psatoken"
)

var (
	genCorimAttestationScheme *string
	genCorimEvidenceFile      *string
	genCorimKeyFile           *string
	genCorimCorimFile         *string
	genCorimTemplateDir       *string
)

var rootCmd = NewRootCmd()

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-corim",
		Short: "generate CoRIM from supplied evidence",
		Long: `generate CoRIM from supplied evidence
		
		Generate CoRIM from evidence token (evidence.cbor), attestation scheme to use (only schemes supported 
		by ths tool are psa and cca) and key material needed to verify the evidence (key.json). 
		Save it to the current working directory with default file name.

				gen-corim --evidence-file=evidence.cbor \
						--key-file=key.json \
						--attest-scheme=scheme

		Generate CoRIM from evidence token (evidence.cbor), attestation scheme to use (only schemes supported 
		by ths tool are psa and cca) and key material needed to verify the evidence (key.json). 
		Save it as target file name (endorsements.cbor)

				gen-corim --evidence-file=evidence.cbor \
						--key-file=key.json \
						--attest-scheme=scheme \
						--corim-file=endorsements.cbor
		`,
		Version: "0.0.1",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkGenCorimArgs(); err != nil {
				return err
			}
			err := generate(genCorimAttestationScheme, genCorimEvidenceFile, genCorimKeyFile, genCorimCorimFile, genCorimTemplateDir)
			if err != nil {
				return err
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	genCorimAttestationScheme = cmd.Flags().StringP("attest-scheme", "a", "", "attestation scheme used")

	genCorimCorimFile = cmd.Flags().StringP("corim-file", "c", "", "name of the generated CoRIM  file")

	genCorimEvidenceFile = cmd.Flags().StringP("evidence-file", "e", "", "a CBOR-encoded evidence file")

	genCorimKeyFile = cmd.Flags().StringP("key-file", "k", "", "a JSON-encoded key file")

	genCorimTemplateDir = cmd.Flags().StringP("template-dir", "t", "", "path of directory containing the comid and corim templates")

	return cmd
}

// checking that the arguments are non-empty and the relevent filepaths exist
func checkGenCorimArgs() error {
	if genCorimAttestationScheme == nil || *genCorimAttestationScheme == "" {
		return errors.New("no attestation scheme supplied")
	}

	if genCorimEvidenceFile == nil || *genCorimEvidenceFile == "" {
		return errors.New("no evidence file supplied")
	}

	if genCorimKeyFile == nil || *genCorimKeyFile == "" {
		return errors.New("no key supplied")
	}

	if *genCorimAttestationScheme != "psa" && *genCorimAttestationScheme != "cca" {
		return errors.New("unsupported attestation scheme, only psa and cca are supported")
	}

	if genCorimTemplateDir == nil || *genCorimTemplateDir == "" {
		return errors.New("no template directory supplied")
	}

	if _, err := os.Stat(*genCorimTemplateDir); errors.Is(err, os.ErrNotExist) {
		return errors.New("template directory does not exist")
	}

	if _, err := os.Stat(*genCorimTemplateDir + "/comid-template.json"); errors.Is(err, os.ErrNotExist) {
		return errors.New("file `comid-template.json` is missing from template directory")
	}

	if _, err := os.Stat(*genCorimTemplateDir + "/corim-template.json"); errors.Is(err, os.ErrNotExist) {
		return errors.New("file `corim-template.json` is missing from template directory")
	}

	return nil
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func generate(attestation_scheme *string, evidence_file *string, key_file *string, corim_file *string, template_dir *string) error {

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error finding working directory: %w", err)
	}

	//creating temporary directory to store intermediate files
	dir, err := os.MkdirTemp(wd, "gen-corim_data")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}

	//validate evidence cryptographically and write to a file
	evcli_cmd := exec.Command("evcli", *attestation_scheme, "check", "--token="+*evidence_file, "--key="+*key_file, "--claims="+dir+"/output-evidence-claims.json")
	if err = evcli_cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error verifying evidence token: %w", err)
	}

	//reading in the evidence token and extracting the claims
	content, err := os.ReadFile(*evidence_file)
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error reading the evidence token: %w", err)
	}

	var claims psatoken.IClaims

	if *attestation_scheme == "psa" {
		var evidence psatoken.Evidence

		err = evidence.FromCOSE(content)
		if err != nil {
			_ = os.RemoveAll(dir)
			return fmt.Errorf("error umarshalling evidence token: %w", err)
		}

		claims = evidence.Claims
	} else {
		var evidence ccatoken.Evidence

		err = evidence.FromCBOR(content)
		if err != nil {
			_ = os.RemoveAll(dir)
			return fmt.Errorf("error umarshalling evidence token: %w", err)
		}

		claims = evidence.PlatformClaims
	}

	//reading in the corim template structure and checking the validity
	content, err = os.ReadFile(*template_dir + "/comid-template.json")
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error reading comid template: %w", err)
	}

	comidClaims := comid.NewComid()
	err = comidClaims.FromJSON(content)
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error umarshalling comid template: %w", err)
	}

	err = comidClaims.Valid()
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error validating comid template: %w", err)
	}

	//storing the key components of the the claims in the desired format
	swComponents, err := claims.GetSoftwareComponents()
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error extracting software components: %w", err)
	}

	implIDByte, err := claims.GetImplID()
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error extracting implementation ID: %w", err)
	}
	var implID comid.ImplID
	copy(implID[:], implIDByte)

	instID, err := claims.GetInstID()
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error extracting instance ID: %w", err)
	}
	var ueid eat.UEID = instID

	//creating a new measurements list to hold the measurements extracted from the evidence token
	measurements := comid.NewMeasurements()

	for _, component := range swComponents {
		refValID := comid.NewPSARefValID(*component.SignerID)
		refValID.SetLabel(*component.MeasurementType)
		refValID.SetVersion(*component.Version)
		measurement := comid.NewPSAMeasurement(*refValID)
		measurement.AddDigest(1, *component.MeasurementValue)
		measurements.AddMeasurement(measurement)
	}

	//adding cca specific measurement
	if *attestation_scheme == "cca" {
		var config, err = claims.GetConfig()
		if err != nil {
			_ = os.RemoveAll(dir)
			return fmt.Errorf("error extracting configuration data: %w", err)
		}
		configID := comid.CCAPlatformConfigID("cfg v1.0.0")
		measurement := comid.NewCCAPlatCfgMeasurement(configID).SetRawValueBytes(config, []byte{})
		measurements.AddMeasurement(measurement)
	}

	//creating a new reference value containing the measurements and the implementation ID from the evidence token
	class := comid.NewClassImplID(implID)

	refVal := comid.ReferenceValue{
		Environment:  comid.Environment{Class: class},
		Measurements: *measurements,
	}

	//replacing the reference values from the template with the created reference value
	referenceValues := append(*new([]comid.ReferenceValue), refVal)
	comidClaims.Triples.ReferenceValues = &referenceValues

	//extracting the key data from the key file and using it to overwrite the AttestVerifKeys triple
	key_data, err := convertJwkToPEM(*key_file)
	if err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	key := comid.NewVerifKey()
	key.SetKey(key_data)
	keys := comid.NewVerifKeys()
	keys.AddVerifKey(key)

	instance := comid.NewInstance()
	instance.SetUEID(ueid)

	verifKey := comid.AttestVerifKey{
		Environment: comid.Environment{
			Class:    class,
			Instance: instance,
		},
		VerifKeys: *keys,
	}

	attestVerifKey := append(*new([]comid.AttestVerifKey), verifKey)
	comidClaims.Triples.AttestVerifKeys = &attestVerifKey

	//writing the constructed claims into a json file to be used as a CoMID template
	content, err = comidClaims.ToJSON()
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error marshalling claims: %w", err)
	}
	os.WriteFile(dir+"/comid-claims.json", content, 0664)

	//creating a CoMID from the constructed template
	comid_cmd := exec.Command("cocli", "comid", "create", "--template="+dir+"/comid-claims.json", "--output-dir="+dir)
	if err := comid_cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error thrown by cocli comid create: %w", err)
	}

	//creating a CoRIM from the CoMID and the provided template
	if *corim_file == "" {
		*corim_file = *attestation_scheme + "-endorsements.cbor"
	}

	corim_cmd := exec.Command("cocli", "corim", "create", "--template="+*template_dir+"/corim-template.json", "--comid="+dir+"/comid-claims.cbor", "--output="+*corim_file)

	if err := corim_cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("error thrown by cocli corim create: %w", err)
	}

	_ = os.RemoveAll(dir)

	return nil
}

func convertJwkToPEM(fileName string) (pemKey string, err error) {
	var buf bytes.Buffer
	// fileName is the name of the file as string type where the JWK is stored
	keyJWK, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("error loading verifying key from %s: %w", fileName, err)
	}
	pkey, err := PubKeyFromJWK(keyJWK)
	if err != nil {
		return "", fmt.Errorf("error loading verifying key from %s: %w", fileName, err)
	}
	pubBytes2, err := x509.MarshalPKIXPublicKey(pkey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes2,
	}
	if err := pem.Encode(&buf, block); err != nil {
		return "", fmt.Errorf("failed to pem encode: %w", err)
	}
	keyStr := buf.String()
	return keyStr, nil
}

// PubKeyFromJWK extracts a crypto.PublicKey from the supplied JSON Web Key
func PubKeyFromJWK(rawJWK []byte) (crypto.PublicKey, error) {
	var pKey crypto.PublicKey
	err := jwk.ParseRawKey(rawJWK, &pKey)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return pKey, nil
}
