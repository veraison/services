package trustedservices

import (
	"github.com/veraison/services/handler"
)

// Interface Component Verifier Client Interface
type ICVClient interface {
	GetCVClient(mt string) (handler.IComponentVerifierClientHandler, error)

	// TO DO, Identify how to get a list of Supported CV Clients..?
	GetSupportedCVClient() ([]handler.IComponentVerifierClientHandler, error)
}

// Dispatch table will be a seperate file which is accessed using dispatch-table key in main config file
// Check issue 369..
