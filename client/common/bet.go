package common

import (
	"bytes"
)

type Bet struct {
	name     string
	surname  string
	dni      string
	birthday string
	number   string
}

func NewBet(name, surname, dni, birthday string, number string) *Bet {
	return &Bet{
		name:     name,
		surname:  surname,
		dni:      dni,
		birthday: birthday,
		number:   number,
	}
}

func (b *Bet) Encode() []byte {
	buf := make([]byte, 0)
	for _, s := range []string{b.name, b.surname, b.dni, b.birthday, b.number} {
		buf = append(buf, byte(uint8(len(s))))
		buf = append(buf, []byte(s)...)
	}
	return buf
}

func Decode(data []byte) *Bet {
	buf := bytes.NewBuffer(data)
	b := &Bet{}
	for _, field := range []*string{&b.name, &b.surname, &b.dni, &b.birthday, &b.number} {
		length, _ := buf.ReadByte()
		*field = string(buf.Next(int(length)))
	}
	return b
}
