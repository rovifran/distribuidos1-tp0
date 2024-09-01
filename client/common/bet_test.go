package common

import (
	"testing"

)

func TestEncodeDecode(t *testing.T) {
	originalBet := NewBet(1, "John", "Pork", 123456789, "1980-01-01", 42)
	encodedBet, err := originalBet.Encode()
	if err != nil {
		t.Errorf("Error encoding bet: %s", err)
	}
	decodedBet, err := Decode(encodedBet)
	if err != nil {
		t.Errorf("Error decoding bet: %s", err)
	}
	if originalBet.name != decodedBet.name {
		t.Errorf("Expected name %s, got %s", originalBet.name, decodedBet.name)
	}
}
