package common

import (
	"bytes"
	"encoding/binary"
)

const SERVER_AMOUNT_OF_BETS_MSG_LEN = 2
const SIZE_UINT32 = 4

type ServerResponse struct {
	AmountOfBets int16
	Winners      []uint32
}

func NewServerResponse(amountOfBets int16, winners []uint32) *ServerResponse {
	return &ServerResponse{
		AmountOfBets: amountOfBets,
		Winners:      winners,
	}
}

// ServerResponseFromBytes Creates a ServerResponse from a byte array
// The byte array must have the following format:
// [2 bytes] - Length of the message in bytes
// [2 bytes] - Result of the operation: Bets processed OR AMount of winners
// [N bytes] - Optional: List of winners. The amount of bytes should be the amount of winners * 4
// because each winner is represented as an uint32
func ServerResponseFromBytes(data []byte) *ServerResponse {
	buf := bytes.NewBuffer(data)
	serverMsgLen := binary.LittleEndian.Uint16(buf.Next(SIZE_UINT16))

	// result can hold either the amount of bets processed or the amount of winners
	result := int16(binary.LittleEndian.Uint16(buf.Next(SIZE_UINT16)))

	// here the handled case is  that the server replies with the amount of bets processed
	// this relies in the precondition that the server can never answer with 0 bets processed
	// because the client can never send 0 bets, and in case of an error, result should be
	// negative
	if serverMsgLen == SERVER_AMOUNT_OF_BETS_MSG_LEN && result != 0 {
		return NewServerResponse(result, nil)
	}

	// the rest of the cases are exclusively the server communicating the amount of winners
	winners := make([]uint32, 0)
	for i := 0; i < int(result); i++{
		winner := binary.LittleEndian.Uint32(buf.Next(SIZE_UINT32)) // the document of the winner
		winners = append(winners, winner)
	}

	return NewServerResponse(0, winners)
}
