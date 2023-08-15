package cmd

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	cogenKeyFile           *string
	cogenAttestationScheme *string
	cogenCorimFile         *string
	cogenEvidenceFile      *string
)

var cogenGenCmd = NewCogenGenCmd()

func NewCogenGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "PLACEHOLDER",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkCogenGenArgs(); err != nil {
				return err
			}
			err := generate(cogenKeyFile, cogenAttestationScheme, cogenCorimFile, cogenEvidenceFile)
			if err != nil {
				return err
			}
			fmt.Printf("PLACEHOLDER")
			return nil
		},
	}

	cogenAttestationScheme = cmd.Flags().StringP("attest-scheme", "a", "", "attestation scheme used")

	cogenCorimFile = cmd.Flags().StringP("corim-files", "c", "", "name of the generated CoRIM  file")

	cogenEvidenceFile = cmd.Flags().StringP("evidence-file", "e", "", "a CBOR-encoded evidence file")

	cogenKeyFile = cmd.Flags().StringP("key-file", "k", "", "a JSON-encoded key file")

	return cmd
}

func checkCogenGenArgs() error {
	if cogenKeyFile == nil || *cogenKeyFile == "" {
		return errors.New("no key supplied")
	}

	if cogenAttestationScheme == nil || *cogenAttestationScheme == "" {
		return errors.New("no attestation scheme supplied")
	}

	if cogenEvidenceFile == nil || *cogenEvidenceFile == "" {
		return errors.New("no evidence file supplied")
	}

	return nil
}

func generate(key_file *string, attestation_scheme *string, corim_file *string, evidence_file *string) error {

	// evcli_cmd := exec.Command("evcli", *attestation_scheme, "check", "--token=", *evidence_file, "--key=", *key_file, "--claims=output-evidence-claims.json")

	// if err := evcli_cmd.Run(); err != nil {
	// 	return err
	// }

	comid_cmd := exec.Command("cocli", "comid", "create", "--template=/home/samdavis/services/utils/coevcli/data/comid-claims.json")

	if err := comid_cmd.Run(); err != nil {
		return err
	}

	//code in here

	// corim_cmd := exec.Command("cocli", "corim create", "--template=comid-template.json", "--comid output-claims.cbor")

	// if *corim_file != "" {
	// 	corim_cmd = exec.Command("cocli", "corim create", "--template=comid-template.json", "--comid output-claims.cbor", "--output", *corim_file)
	// }

	// if err := corim_cmd.Run(); err != nil {
	// 	return err
	// }

	return nil
}

func init() {
	cogenCmd.AddCommand(cogenGenCmd)
}
