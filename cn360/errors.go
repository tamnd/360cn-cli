package cn360

import (
	"errors"
	"fmt"
)

// Exit codes documented in the CLI.
const (
	ExitOK       = 0
	ExitGeneric  = 1
	ExitUsage    = 2
	ExitNotFound = 3
	ExitBlocked  = 5
)

// Sentinel errors returned by the library.
var (
	ErrBlocked     = errors.New("blocked: 360 Search returned a CAPTCHA or anti-bot page")
	ErrRateLimited = errors.New("rate limited: HTTP 429 after retries")
	ErrNotFound    = errors.New("not found")
)

// CodeError carries an exit code alongside a message.
type CodeError struct {
	Code int
	Msg  string
	Err  error
}

func (e *CodeError) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

func (e *CodeError) Unwrap() error { return e.Err }

func codeErr(code int, format string, args ...any) *CodeError {
	return &CodeError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// ExitCode returns the process exit code an error maps to.
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var ce *CodeError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ExitGeneric
}
