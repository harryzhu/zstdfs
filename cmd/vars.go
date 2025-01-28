package cmd

import (
	"errors"
)

const MB int = 1 << 20
const GB int = 1 << 30
const EmptyVal string = ""

var (
	ErrParamEmpty       error = errors.New("args could not be empty")
	ErrParamInvalid     error = errors.New("args invalid")
	ErrNotExist         error = errors.New("does not exist")
	ErrBucketKeyEmpty   error = errors.New("bucket and key cannot be empty")
	ErrFileSizeZero     error = errors.New("filesize is 0")
	ErrFileNameNotMatch error = errors.New("basename of path and relPath are different")
	ErrUnauthorized     error = errors.New("unauthorized")
	ErrAESInvalid       error = errors.New("aes encrypt and decrypt error")
)

var (
	basicAuthFailures int = 0
)
