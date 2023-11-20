package transport

import (
	"errors"
)

var (
  ErrBadFrame   = errors.New("embus: invalid frame")
  ErrOverflow   = errors.New("embus: overflow")
  ErrBoolConvex = errors.New("embus: invalid byte to bool conversion")
)

type Flag byte

func (f Flag) HasMore() bool { return f&SNDMORE == SNDMORE }
func (f Flag) IsLarge() bool { return f&LARGE == LARGE }

// means the peer connector requires a response before sending again
func (f Flag) isSynchronous() bool { return f&SYNC_MSG == SYNC_MSG }

const (
  LARGE    = 0x1
  SNDMORE  = 0x2
  SYNC_MSG = 0x4
)
