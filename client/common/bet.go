package common

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
)

const NAME_POS = 0
const SURNAME_POS = 1
const DNI_POS = 2
const BIRTHDAY_POS = 3
const NUMBER_POS = 4

const AMOUNT_OF_FIELDS = 5

// Struct that encapsulates the bet information, and provides
// methods to encode and decode it
type Bet struct {
	agency   uint8
	name     string
	surname  string
	dni      uint32
	birthday string
	number   uint16
}

// BetReader is a struct that reads the bet information from
// the environment variables and returns a Bet struct
type BetReader struct {
	file      io.Reader
	bufreader *csv.Reader
	maxBets   int
	agency    uint8
	Finished  bool
}

// NewBetReader Initializes a new BetReader
func NewBetReader(maxBets int, agency uint8) *BetReader {
	return &BetReader{
		maxBets: maxBets,
		agency:  agency,
	}
}

// OpenFile Opens the file in the path passed by parameter
func (br *BetReader) OpenFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	br.file = file
	br.bufreader = csv.NewReader(file)
	return nil
}

// CloseFile Closes the file
func (br *BetReader) CloseFile() error {
	if closer, ok := br.file.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// checkFields Checks if the fields of the bet are valid.
// If there are less fields than expected, empty fields, or
// ivalid values for numeric fields, it returns an error
func (br *BetReader) checkFields(line []string) error {
	if len(line) != AMOUNT_OF_FIELDS {
		return errors.New("invalid amount of fields")
	}

	for _, field := range line {
		if len(field) == 0 {
			return errors.New("empty field")
		}
	}

	if _, err := strconv.Atoi(line[DNI_POS]); err != nil {
		return errors.New("invalid DNI")
	}

	if _, err := strconv.Atoi(line[NUMBER_POS]); err != nil {
		return errors.New("invalid number")
	}

	return nil
}

// ReadBets Reads the bet information from the environment variables
// and returns a Bet struct
func (br *BetReader) ReadBets() []*Bet {
	bets := make([]*Bet, 0)
	log.Infof(" maxbets %d", br.maxBets)

	for acum := 0; acum < br.maxBets; acum++ {
		log.Infof("acum: %d", acum)
		line, err := br.bufreader.Read()
		if err != nil {
			if err == csv.ErrFieldCount {
				continue // not enough fields, continue with the next line
			}
			br.Finished = true // if there are this kinds of error, don't read anymore
			if err == io.EOF {
				break
			}
			return nil
		}

		if err = br.checkFields(line); err != nil {
			continue // same case as before
		}

		dni, _ := strconv.Atoi(line[DNI_POS])
		number, _ := strconv.Atoi(line[NUMBER_POS])

		bet := NewBet(br.agency, line[NAME_POS], line[SURNAME_POS], uint32(dni), line[BIRTHDAY_POS], uint16(number))
		bets = append(bets, bet)
	}

	return bets
}

// NewBet Initializes a new Bet struct given the corresponding parameters
func NewBet(agency uint8, name string, surname string, dni uint32, birthday string, number uint16) *Bet {
	return &Bet{
		agency:   agency,
		name:     name,
		surname:  surname,
		dni:      dni,
		birthday: birthday,
		number:   number,
	}
}

// Encode Encodes the bet information into a byte array. If an error
// occurs, it is returned
func (b *Bet) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	agencyBytes := make([]byte, 1)
	agencyBytes[0] = b.agency
	_, err := SafeWriteBytes(buf, agencyBytes)
	if err != nil {
		return nil, err
	}

	_, err = SafeWriteStringField(buf, b.name)
	if err != nil {
		return nil, err
	}

	_, err = SafeWriteStringField(buf, b.surname)
	if err != nil {
		return nil, err
	}

	dniBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(dniBytes, b.dni)
	_, err = SafeWriteBytes(buf, dniBytes)
	if err != nil {
		return nil, err
	}

	_, err = SafeWriteStringField(buf, b.birthday)
	if err != nil {
		return nil, err
	}

	numberBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(numberBytes, b.number)
	_, err = SafeWriteBytes(buf, numberBytes)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// createStrFieldFromBytes Creates a string field from a byte array
// by reading the length of the field and then the field itself.
// If an error occurs, it is returned
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

// Decode Decodes a byte array into a Bet struct. If an error occurs,
// it is returned
func Decode(data []byte) (*Bet, error) {
	buf := bytes.NewBuffer(data)
	agency := buf.Next(1)[0]
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

	return NewBet(agency, name, surname, dni, birthday, number), nil

}

// EncodeBets Encodes a slice of bets into a byte array. If an error
// occurs, it is returned
func EncodeBets(bets []*Bet) ([]byte, error) {
	buf := new(bytes.Buffer)

	for _, bet := range bets {
		encodedBet, err := bet.Encode()
		if err != nil {
			return nil, err
		}

		_, err = SafeWriteBytes(buf, []byte{uint8(len(encodedBet))})
		if err != nil {
			return nil, err
		}

		_, err = SafeWriteBytes(buf, encodedBet)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
