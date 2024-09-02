package common

import "encoding/binary"

type ServerResponse struct {
	AmountOfBets int16
}

func NewServerResponse(amountOfBets int16) *ServerResponse {
	return &ServerResponse{
		AmountOfBets: amountOfBets,
	}
}

func decode(bytes []byte) int16 {
	return int16(binary.LittleEndian.Uint16(bytes))
}

func ServerResponseFromBytes(bytes []byte) *ServerResponse {
	amountOfBets := decode(bytes)
	return &ServerResponse{
		AmountOfBets: amountOfBets,
	}
}
