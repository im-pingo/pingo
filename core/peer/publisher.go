package peer

import (
	"github.com/pingostack/pingos/core/mode"
)

type Publisher interface {
	mode.ReadCloser
	// ID returns the unique identifier of the publisher
	ID() string
	Format() mode.FmtType
	AddWriter(writer mode.WriteCloser) error
}
