package rmqrpc

import "errors"

var (
	// ErrTimeout -.
	ErrTimeout = errors.New("timeout")
	// ErrInternalServer -.
	ErrInternalServer = errors.New("core server error")
	// ErrBadHandler -.
	ErrBadHandler = errors.New("unregistered handler")
)

// Success -.
const Success = "success"
