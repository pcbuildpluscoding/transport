package flow

import (
	"errors"

	logg "github.com/pcbuildpluscoding/logroll"
)

var logger = logg.New()

// -------------------------------------------------------------- //
// SetLogger
// ---------------------------------------------------------------//
func SetLogger(super *logg.LogFile) {
	logger = super
}

var (
	ErrBadFrame = errors.New("flowware: invalid frame format")
	ErrOverflow = errors.New("flowware: size overflows uint16 maximum")
)

type Flag byte

func (f Flag) HasMore() bool { return f&SNDMORE == SNDMORE }

const (
	SNDMORE = 1
)
