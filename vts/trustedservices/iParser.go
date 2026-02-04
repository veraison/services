package trustedservices

import cp "github.com/veraison/services/vts/compositeevidenceparser"

// Interface parser
type IParser interface {
	GetParserFromMediaType(mt string) (cp.ICompositeEvidenceParser, error)

	// TO DO, Identify how to get a list of Supported Parsers under Veraison..?
	GetSupportedParsers() ([]cp.ICompositeEvidenceParser, error)
}
