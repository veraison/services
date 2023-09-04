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
	"github.com/spf13/viper"
	"github.com/veraison/ccatoken"
	"github.com/veraison/corim/comid"
	"github.com/veraison/eat"
	"github.com/veraison/psatoken"
)

var (
	cfgFile                  string
	coevcliAttestationScheme *string
	coevcliEvidenceFile      *string
	coevcliKeyFile           *string
	coevcliCorimFile         *string
)

var rootCmd = NewRootCmd()

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "coevcli",
		Short:   "create corim from supplied evidence",
		Version: "0.0.1",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkCogenGenArgs(); err != nil {
				return err
			}
			err := generate(coevcliAttestationScheme, coevcliEvidenceFile, coevcliKeyFile, coevcliCorimFile)
			if err != nil {
				return err
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	coevcliAttestationScheme = cmd.Flags().StringP("attest-scheme", "a", "", "attestation scheme used")

	coevcliCorimFile = cmd.Flags().StringP("corim-file", "c", "", "name of the generated CoRIM  file")

	coevcliEvidenceFile = cmd.Flags().StringP("evidence-file", "e", "", "a CBOR-encoded evidence file")

	coevcliKeyFile = cmd.Flags().StringP("key-file", "k", "", "a JSON-encoded key file")

	return cmd
}

func checkCogenGenArgs() error {
	if coevcliAttestationScheme == nil || *coevcliAttestationScheme == "" {
		return errors.New("no attestation scheme supplied")
	}

	if coevcliEvidenceFile == nil || *coevcliEvidenceFile == "" {
		return errors.New("no evidence file supplied")
	}

	if coevcliKeyFile == nil || *coevcliKeyFile == "" {
		return errors.New("no key supplied")
	}

	if *coevcliAttestationScheme != "psa" && *coevcliAttestationScheme != "cca" {
		return errors.New("unsupported attestation scheme")
	}

	return nil
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// search config in home directory with name ".cli" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func generate(attestation_scheme *string, evidence_file *string, key_file *string, corim_file *string) error {

	evcli_cmd := exec.Command("evcli", *attestation_scheme, "check", "--token="+*evidence_file, "--key="+*key_file, "--claims=../data/output-evidence-claims.json")
	if err := evcli_cmd.Run(); err != nil {
		return err
	}

	content, err := os.ReadFile(*evidence_file)
	if err != nil {
		return err
	}

	var claims psatoken.IClaims

	if *attestation_scheme == "psa" {
		var evidence psatoken.Evidence

		err = evidence.FromCOSE(content)
		if err != nil {
			return err
		}

		claims = evidence.Claims
	} else {
		var evidence ccatoken.Evidence

		err = evidence.FromCBOR(content)
		if err != nil {
			return err
		}

		claims = evidence.PlatformClaims
	}

	swComponents, err := claims.GetSoftwareComponents()
	if err != nil {
		return err
	}

	implIDByte, err := claims.GetImplID()
	if err != nil {
		return err
	}
	var implID comid.ImplID
	copy(implID[:], implIDByte)

	instID, err := claims.GetInstID()
	if err != nil {
		return err
	}
	var ueid eat.UEID = instID

	measurements := comid.NewMeasurements()

	for _, component := range swComponents {
		refValID := comid.NewPSARefValID(*component.SignerID)
		refValID.SetLabel(*component.MeasurementType)
		refValID.SetVersion(*component.Version)
		measurement := comid.NewPSAMeasurement(*refValID)
		measurement.AddDigest(1, *component.MeasurementValue)
		measurements.AddMeasurement(measurement)
	}

	if *attestation_scheme == "cca" {
		var config, err = claims.GetConfig()
		if err != nil {
			return err
		}
		configID := comid.CCAPlatformConfigID("cfg v1.0.0")
		measurement := comid.NewCCAPlatCfgMeasurement(configID).SetRawValueBytes(config, []byte{})
		measurements.AddMeasurement(measurement)
	}

	class := comid.NewClassImplID(implID)

	refVal := comid.ReferenceValue{
		Environment:  comid.Environment{Class: class},
		Measurements: *measurements,
	}

	content, err = os.ReadFile("../data/comid-claims-template.json")
	if err != nil {
		return err
	}

	comidClaims := comid.NewComid()
	err = comidClaims.FromJSON(content)
	if err != nil {
		return err
	}

	referenceValues := append(*new([]comid.ReferenceValue), refVal)
	comidClaims.Triples.ReferenceValues = &referenceValues

	key_data, err := convertJwkToPEM(*key_file)
	if err != nil {
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

	content, err = comidClaims.ToJSON()
	if err != nil {
		return err
	}
	os.WriteFile("../data/comid-claims.json", content, 0664)

	comid_cmd := exec.Command("cocli", "comid", "create", "--template=../data/comid-claims.json", "--output-dir=../data")
	if err := comid_cmd.Run(); err != nil {
		return err
	}

	corim_cmd := exec.Command("cocli", "corim", "create", "--template=../data/corim-full.json", "--comid=../data/comid-claims.cbor", "--output=../data/"+*attestation_scheme+"-endorsements.cbor")

	if *corim_file != "" {
		corim_cmd = exec.Command("cocli", "corim", "create", "--template=../data/corim-full.json", "--comid=../data/comid-claims.cbor", "--output="+*corim_file)
	}

	if err := corim_cmd.Run(); err != nil {
		return err
	}

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
