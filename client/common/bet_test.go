package common

import (
	"fmt"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	originalBet := NewBet(1, "John", "Pork", 123456789, "1980-01-01", 42)
	encodedBet, err := originalBet.Encode()
	fmt.Println(encodedBet)
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

func TestBetReaderIsNotFinishedWhenMoreBetsCanBeRead(t *testing.T) {
	betReader := BetReader{
		maxBets: 15,
		agency:  0,
	}

	betReader.OpenFile("../test_files/bets.csv")
	defer betReader.CloseFile()

	bets := betReader.ReadBets()
	if len(bets) != 15 {
		t.Errorf("Expected 15 bets, got %d instead", len(bets))
	}
	if betReader.Finished {
		t.Error("Bet reader should not be finished")
	}
}

func TestBetReaderIsFinishedWhenNoMoreBetsCanBeRead(t *testing.T) {
	betReader := BetReader{
		maxBets: 60,
		agency:  0,
	}

	betReader.OpenFile("../test_files/bets.csv")
	defer betReader.CloseFile()

	if bets := betReader.ReadBets(); len(bets) != 50 {
		t.Errorf("Expected 50 bets, got %d instead", len(bets))
	}

	if !betReader.Finished {
		t.Error("Bet reader should be finished")
	}
}
