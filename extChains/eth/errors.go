package eth

import "errors"

var (
	ErrClientUrlsEmpty  = errors.New("empty client urls")
	ErrEcdsaPubCast     = errors.New("can not cast public key to ECDSA")
	ErrUnsupportedEvent = errors.New("unsupported event")
	ErrHandlerUndefined = errors.New("request handler not defined")
)
