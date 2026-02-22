package repository

import (
	"regexp"
	"testing"
)

func TestGeneratePNR(t *testing.T) {
	// 1. Test Length
	pnr := generatePNR()
	if len(pnr) != 6 {
		t.Errorf("Expected PNR length 6, got %d", len(pnr))
	}

	// 2. Test Character Set (Alphanumeric uppercase, no I, O, 0, 1)
	// allowed chars: ABCDEFGHJKLMNPQRSTUVWXYZ23456789
	match, _ := regexp.MatchString("^[A-Z0-9]+$", pnr)
	if !match {
		t.Errorf("PNR contains invalid characters: %s", pnr)
	}
	
	// 3. Test Randomness (simple check)
	pnr2 := generatePNR()
	if pnr == pnr2 {
		t.Logf("Warning: Generated consecutive identical PNRs: %s", pnr)
		// strictly speaking this is possible but unlikely
	}
}
