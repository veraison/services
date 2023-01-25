package builtin

import (
	"github.com/veraison/services/plugin"

	scheme1 "github.com/veraison/services/scheme/cca-ssd-platform"
	scheme4 "github.com/veraison/services/scheme/psa-iot"
	scheme2 "github.com/veraison/services/scheme/tcg-dice"
	scheme3 "github.com/veraison/services/scheme/tpm-enacttrust"
)

var plugins = []plugin.IPluggable{
	&scheme1.EvidenceDecoder{},
	&scheme1.EndorsementDecoder{},
	&scheme2.EvidenceDecoder{},
	&scheme3.EvidenceDecoder{},
	&scheme3.EndorsementDecoder{},
	&scheme4.EvidenceDecoder{},
	&scheme4.EndorsementDecoder{},
}
