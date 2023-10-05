package builtin

import (
	"github.com/veraison/services/plugin"

	scheme1 "github.com/veraison/services/scheme/cca-ssd-platform"
	scheme4 "github.com/veraison/services/scheme/psa-iot"
	scheme2 "github.com/veraison/services/scheme/riot"
	scheme3 "github.com/veraison/services/scheme/tpm-enacttrust"
)

var plugins = []plugin.IPluggable{
	&scheme1.EvidenceHandler{},
	&scheme1.EndorsementHandler{},
	&scheme2.EvidenceHandler{},
	&scheme3.EvidenceHandler{},
	&scheme3.EndorsementHandler{},
	&scheme4.EvidenceHandler{},
	&scheme4.EndorsementHandler{},
}
