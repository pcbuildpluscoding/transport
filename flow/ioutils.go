package flow

import (
	"encoding/binary"
	"io"
)

// ----------------------------------------------------------------//
// newHeader
// ----------------------------------------------------------------//
func newHeader1(frame []byte, flag byte) []byte {
	hdr := make([]byte, 3)
	hdr[0] = flag
	binary.LittleEndian.PutUint16(hdr[1:], uint16(len(frame)))

	return hdr
}

// ----------------------------------------------------------------//
// Read
// ----------------------------------------------------------------//
func ReadAll(r io.Reader, frame []byte) error {
	var (
		n   int
		err error
	)
	for remnant := len(frame); remnant > 0; remnant -= n {
		n, err = r.Read(frame)
		if err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------//
// ReadWH - read header first, get frame length, read frame
// ----------------------------------------------------------------//
func ReadWH2(r io.Reader) (Flag, []byte, error) {
	var (
		header [3]byte
		err    error
	)
	// read the header
	err = ReadAll(r, header[:])
	if err != nil {
		return 0, nil, err
	}

	flag := Flag(header[0])
	// Determine the actual length of the body
	fsize := binary.LittleEndian.Uint16(header[1:])

	frame := make([]byte, fsize)
	err = ReadAll(r, frame)

	return flag, frame, err
}

// ----------------------------------------------------------------//
// ReadPacket
// ----------------------------------------------------------------//
func ReadPacket2(r io.Reader) ([][]byte, error) {
	flag, frame, err := ReadWH2(r)
	if err != nil {
		return nil, err
	}
	packet := [][]byte{frame}
	for flag.HasMore() {
		// read the header
		flag, frame, err = ReadWH2(r)
		if err != nil {
			return nil, err
		}
		packet = append(packet, frame)
	}
	return packet, nil
}
