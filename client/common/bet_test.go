package common

import (
	"testing"

)

func TestEncodeDecode(t *testing.T) {
	originalBet := NewBet("John", "Pork", "123456789", "1980-01-01", "42")
	encodedBet := originalBet.Encode()
	decodedBet := Decode(encodedBet)
	if originalBet.name != decodedBet.name {
		t.Errorf("Expected name %s, got %s", originalBet.name, decodedBet.name)
	}
}
