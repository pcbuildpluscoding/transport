package transport

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"syscall"
	"time"
)

// ----------------------------------------------------------------//
// newHeader
// ----------------------------------------------------------------//
func newHeader(frame []byte, flag byte) []byte {
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
// Read1
// ----------------------------------------------------------------//
func Read1(c net.Conn, frame []byte) error {
	var (
		n   int
		err error
	)
	retries := 3
	dura := time.Duration(2) * time.Second
	for remnant := len(frame); remnant > 0; remnant -= n {
		if remnant > 16384 {
			remnant = 16384
		}
		err = c.SetReadDeadline(time.Now().Add(dura))
		if err != nil {
			return err
		}
		n, err = c.Read(frame)
		if err != nil {
			if err == syscall.EAGAIN || err == os.ErrDeadlineExceeded {
				if retries == 0 {
					return os.ErrDeadlineExceeded
				}
				retries--
				continue
			}
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------//
// ReadWH - read header first, get frame length, read frame
// ----------------------------------------------------------------//
func ReadWH1(c net.Conn) (Flag, []byte, error) {
	var (
		header [2]byte
		extHdr [4]byte
		err    error
	)
	// read the header
	err = Read1(c, header[:])
	if err != nil {
		return 0, nil, err
	}

	flag := Flag(header[0])
	// Determine the actual length of the body
	fsize := uint32(header[1])
	// Determine the actual length of the body
	if flag.IsLarge() {
		extHdr[0] = header[1]
		err = Read1(c, extHdr[1:])
		if err != nil {
			if err == syscall.EAGAIN {
				return 0, nil, nil
			}
			return 0, nil, err
		}
		fsize = binary.BigEndian.Uint32(extHdr[:])
	}

	if fsize > uint32(math.MaxUint32) {
		return 0, nil, ErrOverflow
	}

	frame := make([]byte, fsize)
	err = Read1(c, frame)
	return flag, frame, err
}

// ----------------------------------------------------------------//
// ReadPacket
// ----------------------------------------------------------------//
func ReadPacket1(c net.Conn) ([][]byte, error) {
	flag, frame, err := ReadWH1(c)
	if err != nil {
		return nil, err
	}
	packet := [][]byte{frame}
	for flag.HasMore() {
		// read the header
		flag, frame, err = ReadWH1(c)
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
func Write1(c net.Conn, frame []byte) error {
	var (
		n   int
		err error
	)

	retries := 3
	dura := time.Duration(2) * time.Second
	for total := 0; total < len(frame); total += n {
		err = c.SetWriteDeadline(time.Now().Add(dura))
		if err != nil {
			return err
		}
		n, err = c.Write(frame[total:])
		if err != nil {
			if err == syscall.EAGAIN || err == os.ErrDeadlineExceeded {
				if retries == 0 {
					return os.ErrDeadlineExceeded
				}
				retries--
				continue
			}
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------//
// WriteWH - write with header
// ----------------------------------------------------------------//
func WriteWH1(c net.Conn, frame []byte, flag byte) error {
	header := newHeader(frame, flag)
	err := Write1(c, header)
	if err != nil {
		return err
	}

	return Write1(c, frame)
}

// ----------------------------------------------------------------//
// WriteHPacket
// ----------------------------------------------------------------//
func WriteHPacket1(c net.Conn, header []byte, packet ...[]byte) error {
	var flag byte
	if len(packet) > 0 {
		flag ^= SNDMORE
	}
	err := WriteWH1(c, header, flag)
	if err != nil {
		return fmt.Errorf("failed sending header : %w", err)
	}
	return WritePacket1(c, packet...)
}

// ----------------------------------------------------------------//
// WritePacket without a msgtype header
// ----------------------------------------------------------------//
func WritePacket1(c net.Conn, packet ...[]byte) error {
	var flag byte
	last := len(packet) - 1
	for i, frame := range packet {
		flag = SNDMORE
		if i == last {
			flag = 0
		}
		err := WriteWH1(c, frame, flag)
		if err != nil {
			return fmt.Errorf("failed sending frame %d of %d : %w", i+1, last+1, err)
		}
	}
	return nil
}
