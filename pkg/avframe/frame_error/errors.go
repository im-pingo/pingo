package frameerror

import (
	"errors"
	"io"
)

var ErrDecline = errors.New("decline frame")
var ErrBreak = errors.New("break frame process")
var ErrEOF = io.EOF
var ErrAgain = errors.New("again")
