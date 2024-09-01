package common

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strconv"
)

type Bet struct {
	name     string
	surname  string
	dni      uint32
	birthday string
	number   uint16
}

type BetReader struct{}

func NewBetReader() *BetReader {
	return &BetReader{}
}

func (br *BetReader) ReadBets() *Bet {
	dni, _ := strconv.ParseUint(os.Getenv("dni"), 10, 32)
	number, _ := strconv.ParseUint(os.Getenv("numero"), 10, 16)
	bet := NewBet(
		os.Getenv("nombre"),
		os.Getenv("apellido"),
		uint32(dni),
		os.Getenv("nacimiento"),
		uint16(number),
	)
	return bet
}

func NewBet(name string, surname string, dni uint32, birthday string, number uint16) *Bet {
	return &Bet{
		name:     name,
		surname:  surname,
		dni:      dni,
		birthday: birthday,
		number:   number,
	}
}

func (b *Bet) safeWriteBytes(buf *bytes.Buffer, bytes []byte) (int, error) {
	writtenBytes := 0
	for writtenBytes < len(bytes) {
		n, err := buf.Write(bytes[writtenBytes:])
		if err != nil && err != io.ErrShortWrite {
			return 0, err
		}
		writtenBytes += n
	}
	return writtenBytes, nil
}

func (b *Bet) safeWriteStringField(buf *bytes.Buffer, field string) (int, error) {
	encodedFieldLen := byte(uint8(len(field)))
	encodedField := []byte(field)

	for _, field := range [][]byte{[]byte{encodedFieldLen}, encodedField} {
		_, err := b.safeWriteBytes(buf, field)
		if err != nil {
			return 0, err
		}
	}
	return len(encodedField) + 1, nil
}
func (b *Bet) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := b.safeWriteStringField(buf, b.name)
	if err != nil {
		return nil, err
	}

	_, err = b.safeWriteStringField(buf, b.surname)
	if err != nil {
		return nil, err
	}

	dniBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(dniBytes, b.dni)
	_, err = b.safeWriteBytes(buf, dniBytes)
	if err != nil {
		return nil, err
	}

	_, err = b.safeWriteStringField(buf, b.birthday)
	if err != nil {
		return nil, err
	}

	numberBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(numberBytes, b.number)
	_, err = b.safeWriteBytes(buf, numberBytes)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func createStrFieldFromBytes(buf *bytes.Buffer) (string, error) {
	fieldLen, err := buf.ReadByte()
	if fieldLen == 0 || err != nil {
		return "", err
	}

	field := make([]byte, int(fieldLen))
	bytesRead := 0
	for bytesRead < int(fieldLen) {
		n, err := buf.Read(field[bytesRead:])
		if err != nil && err != io.EOF {
			return "", err
		}
		bytesRead += n
	}

	return string(field), nil
}

func Decode(data []byte) (*Bet, error) {
	buf := bytes.NewBuffer(data)
	name, err := createStrFieldFromBytes(buf)
	if err != nil {
		return nil, err
	}
	surname, err := createStrFieldFromBytes(buf)
	if err != nil {
		return nil, err
	}
	dni := binary.LittleEndian.Uint32(buf.Next(4))
	birthday, err := createStrFieldFromBytes(buf)
	if err != nil {
		return nil, err
	}
	number := binary.LittleEndian.Uint16(buf.Next(2))

	return NewBet(name, surname, dni, birthday, number), nil

}
