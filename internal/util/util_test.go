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
	"testing"
)

func TestIsDisconnection(t *testing.T) {
	cases := []struct {
		err error
		expected bool
	}{
		{nil, false},
		{errors.New("foobar"), false},
		{io.EOF, true},
		{syscall.EPIPE, true},
		{syscall.ECONNRESET, true},
		{syscall.ETIMEDOUT, true},
		{os.ErrDeadlineExceeded, true},
	}

	for _, tst := range cases {
		actual := IsDisconnection(tst.err)
		if actual != tst.expected {
			t.Errorf("unexpected disconnected truth value: %t (expecting %t) for case: %v", actual, tst.expected, tst.err)
		}
	}
}
