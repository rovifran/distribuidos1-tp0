package common

import (
	"bytes"
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

func SafeReadBytes(buf io.Reader, responseBytes []byte) (int, error) {
	readBytes := 0
	for readBytes < SERVER_MSG_SIZE {
		n, err := buf.Read(responseBytes[readBytes:])
		if err != nil && err != io.EOF {
			return 0, err
		}
		readBytes += n
	}

	return readBytes, nil
}