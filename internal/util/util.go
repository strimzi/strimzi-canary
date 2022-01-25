//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package util contains some utility functions
package util

import (
	"errors"
	"io"
	"os"
	"syscall"
	"time"
)

// NowInMilliseconds returns the current time in milliseconds
func NowInMilliseconds() int64 {
	return time.Now().UnixNano() / 1000000
}

// IsDisconnection returns true if the err provided represents a TCP disconnection
func IsDisconnection(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ETIMEDOUT) || errors.Is(err, os.ErrDeadlineExceeded)
}
