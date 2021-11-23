package extChains

import "errors"

var (
	ErrUnsupported               = errors.New("unsupported now")
	ErrClientIdMismatchToNetwork = errors.New("mismatch to network")
)
