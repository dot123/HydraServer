package errors

import "github.com/pkg/errors"

// Define alias
var (
	Is           = errors.Is
	New          = errors.New
	As           = errors.As
	Wrap         = errors.Wrap
	Wrapf        = errors.Wrapf
	WithStack    = errors.WithStack
	WithMessage  = errors.WithMessage
	WithMessagef = errors.WithMessagef
)
