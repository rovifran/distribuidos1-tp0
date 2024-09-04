package common

import (
	"bytes"
	"encoding/binary"
	"io"
)

// SafeWriteBytes Writes bytes to a buffer, ensuring that all bytes are written
// thus tolerating short writes: keep writing until all bytes are written
// or, if an error occurs and it is not a short write, return the error
func SafeWriteBytes(buf io.Writer, bytes []byte) (int, error) {
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

// SafeWriteStringField Writes a string field to a buffer, as well as
// its length
func SafeWriteStringField(buf *bytes.Buffer, field string) (int, error) {
	encodedFieldLen := byte(uint8(len(field)))
	encodedField := []byte(field)

	for _, field := range [][]byte{{encodedFieldLen}, encodedField} {
		_, err := SafeWriteBytes(buf, field)
		if err != nil {
			return 0, err
		}
	}
	return len(encodedField) + 1, nil
}

// SafeReadBytes Reads bytes from a buffer, ensuring that all bytes are read
// thus tolerating short reads: keep reading until all bytes are written in the buffer
// or, if an error occurs and it is not a short read, return the error
func SafeReadBytes(buf io.Reader, responseBytes []byte, bytesToRead int) (int, error) {
	readBytes := 0
	for readBytes < bytesToRead {
		n, err := buf.Read(responseBytes[readBytes:])
		if n == 0 && err == io.EOF {
			break
		}

		if err != nil && err != io.EOF {
			return 0, err
		}
		readBytes += n
	}

	return readBytes, nil
}

func SafeReadVariableBytes(buf io.Reader, responseBytes []byte) ([]byte, error) {
	readBytes := 0
	msgLenBytes := make([]byte, LEN_SERVER_MSG_SIZE)
	n, err := SafeReadBytes(buf, msgLenBytes, LEN_SERVER_MSG_SIZE)
	if n == 0 && err == io.EOF {
		return make([]byte, 0), err
	}

	if err != nil {
		return nil, err
	}

	readBytes += n
	msglen := binary.LittleEndian.Uint16(msgLenBytes)
	msgBytes := make([]byte, msglen)
	res := make([]byte, msglen)
	n, err = SafeReadBytes(buf, msgBytes, int(msglen))
	if n == 0 && err == io.EOF {
		return make([]byte, 0), err
	}
	if err != nil {
		return nil, err
	}

	copy(res, msgBytes)

	readBytes += n
	return res, nil
}
