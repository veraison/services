package proto

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func Test_TrustTier_String(t *testing.T) {
	assert.Equal(t, TrustTier_AFFIRMING.String(), "AFFIRMING")
	assert.Equal(t, TrustTier(73).String(), "TrustTier(73)")
}

func Test_ARStatus_GetTier(t *testing.T) {
	assert.Equal(t, ARStatus(-1).GetTier(), TrustTier_NONE)
	assert.Equal(t, ARStatus(0).GetTier(), TrustTier_NONE)
	assert.Equal(t, ARStatus(1).GetTier(), TrustTier_NONE)
	assert.Equal(t, ARStatus(-2).GetTier(), TrustTier_AFFIRMING)
	assert.Equal(t, ARStatus(10).GetTier(), TrustTier_AFFIRMING)
	assert.Equal(t, ARStatus(-40).GetTier(), TrustTier_WARNING)
	assert.Equal(t, ARStatus(95).GetTier(), TrustTier_WARNING)
	assert.Equal(t, ARStatus(-128).GetTier(), TrustTier_CONTRAINDICATED)
	assert.Equal(t, ARStatus(101).GetTier(), TrustTier_CONTRAINDICATED)
}
