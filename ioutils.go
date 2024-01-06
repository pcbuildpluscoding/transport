// The MIT License
//
// Copyright (c) 2020 Peter A McGill
package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"syscall"
)

var ErrConnexCancelled = errors.New("network connection was cancelled")

// ----------------------------------------------------------------//
// Closure
// ----------------------------------------------------------------//
func Closure(err error) bool {
	netErr, isNetErr := err.(*net.OpError)
	if isNetErr {
		errno, isErrno := netErr.Err.(syscall.Errno)
		return isErrno && errno == syscall.ECONNRESET
	}
	return false
}

// ----------------------------------------------------------------//
// GetReply
// ----------------------------------------------------------------//
func GetReply(r io.Reader) *ApiNote {
	status := &ApiNote{}
	_, frame, err := ReadWH(r)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return status.Withf(500, "io.Reader failed : peer connection closed")
		default:
			logger.Errorf("io.Reader failed : %v", err)
			return status.With(400, err)
		}
	}
	err = status.Decode(frame)
	if err != nil {
		return status.With(400, err)
	}
	// logger.Debugf("GetReply response : %v", status)
	return status
}

// ----------------------------------------------------------------//
// ParseHeader
// ----------------------------------------------------------------//
func ParseHeader(frame []byte, flag byte) []byte {
	// Long flag
	fsize := len(frame)
	large := fsize > 255
	if large {
		flag ^= LARGE
	}

	var (
		// flag has position 0
		header = [5]byte{flag}
		hsize  int
	)

	if large {
		hsize = 5
		binary.BigEndian.PutUint32(header[1:], uint32(fsize))
	} else {
		hsize = 2
		header[1] = byte(fsize)
	}
	return header[:hsize]
}

// ----------------------------------------------------------------//
// Read
// ----------------------------------------------------------------//
func Read(r io.Reader, frame []byte) error {
	var (
		n   int
		err error
	)
	for remnant := len(frame); remnant > 0; remnant -= n {
		if remnant > 16384 {
			remnant = 16384
		}
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
func ReadWH(r io.Reader) (Flag, []byte, error) {
	var (
		header [2]byte
		extHdr [4]byte
		err    error
	)
	// read the header
	err = Read(r, header[:])
	if err != nil {
		return 0, nil, err
	}

	flag := Flag(header[0])
	// Determine the actual length of the body
	fsize := uint32(header[1])
	// Determine the actual length of the body
	if flag.IsLarge() {
		extHdr[0] = header[1]
		err = Read(r, extHdr[1:])
		if err != nil {
			return 0, nil, err
		}
		fsize = binary.BigEndian.Uint32(extHdr[:])
	}

	if fsize > uint32(math.MaxUint32) {
		return 0, nil, ErrOverflow
	}

	frame := make([]byte, fsize)
	err = Read(r, frame)

	if err != nil {
		return 0, nil, err
	}
	return flag, frame, nil
}

// ----------------------------------------------------------------//
// ReadPacket
// ----------------------------------------------------------------//
func ReadPacket(r io.Reader) ([][]byte, error) {
	flag, frame, err := ReadWH(r)
	if err != nil {
		return nil, err
	}
	packet := [][]byte{frame}
	for flag.HasMore() {
		// read the header
		flag, frame, err = ReadWH(r)
		if err != nil {
			return nil, err
		}
		packet = append(packet, frame)
	}
	return packet, nil
}

// ----------------------------------------------------------------//
// Write
// ----------------------------------------------------------------//
func Write(w io.Writer, frame []byte) error {
	var (
		n   int
		err error
	)

	for total := 0; total < len(frame); total += n {
		n, err = w.Write(frame[total:])
		if err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------//
// WriteWH - write with header
// ----------------------------------------------------------------//
func WriteWH(w io.Writer, frame []byte, flag byte) error {
	header := ParseHeader(frame, flag)
	err := Write(w, header)
	if err != nil {
		return err
	}

	return Write(w, frame)
}

// ----------------------------------------------------------------//
// WriteHPacket
// ----------------------------------------------------------------//
func WriteHPacket(w io.Writer, header []byte, packet ...[]byte) error {
	var flag byte
	if len(packet) > 0 {
		flag ^= SNDMORE
	}
	err := WriteWH(w, header, flag)
	if err != nil {
		return fmt.Errorf("failed sending header : %w", err)
	}
	return WritePacket(w, packet...)
}

// ----------------------------------------------------------------//
// WritePacket without a msgtype header
// ----------------------------------------------------------------//
func WritePacket(w io.Writer, packet ...[]byte) error {
	var flag byte
	last := len(packet) - 1
	for i, frame := range packet {
		flag = SNDMORE
		if i == last {
			flag = 0
		}
		err := WriteWH(w, frame, flag)
		if err != nil {
			return fmt.Errorf("failed sending frame %d of %d : %w", i+1, last+1, err)
		}
	}
	return nil
}
