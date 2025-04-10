package builtin

import (
	"github.com/veraison/services/plugin"

	scheme3 "github.com/veraison/services/scheme/arm-cca"
	scheme1 "github.com/veraison/services/scheme/parsec-cca"
	scheme5 "github.com/veraison/services/scheme/parsec-tpm"
	scheme6 "github.com/veraison/services/scheme/psa-iot"
	scheme2 "github.com/veraison/services/scheme/riot"
	scheme4 "github.com/veraison/services/scheme/tpm-enacttrust"
)

var plugins = []plugin.IPluggable{
	&scheme1.EvidenceHandler{},
	&scheme1.EndorsementHandler{},
	&scheme1.StoreHandler{},
	&scheme2.EvidenceHandler{},
	&scheme2.StoreHandler{},
	&scheme3.EvidenceHandler{},
	&scheme3.EndorsementHandler{},
	&scheme3.StoreHandler{},
	&scheme4.EvidenceHandler{},
	&scheme4.EndorsementHandler{},
	&scheme4.StoreHandler{},
	&scheme5.EvidenceHandler{},
	&scheme5.EndorsementHandler{},
	&scheme5.StoreHandler{},
	&scheme6.EvidenceHandler{},
	&scheme6.EndorsementHandler{},
	&scheme6.StoreHandler{},
}
