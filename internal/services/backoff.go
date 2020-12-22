//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import (
	"fmt"
	"time"
)

const (
	ScaleDefault       = 200
	BaseDefault        = 2
	MaxAttemptsDefault = 6
)

// Backoff encapsulates computing delays for an exponential back-off, when an operation has to be retried
type Backoff struct {
	maxAttempts int
	scale       int64
	base        int
	attempt     int
}

// MaxAttemptsExceeded defines the error for the max attempts exceeded
type MaxAttemptsExceeded struct{}

func (e *MaxAttemptsExceeded) Error() string {
	return fmt.Sprintf("Maximum number of attempts exceeded")
}

// NewBackoff returns an instance of a Backoff struct
func NewBackoff(maxAttempts int, scale int64, base int) *Backoff {
	backoff := Backoff{
		maxAttempts: maxAttempts,
		scale:       scale,
		base:        base,
		attempt:     0,
	}
	return &backoff
}

// Delay computes a delay in milliseconds based on the current Backoff instance configuration
// Returns the delay in milliseconds and an error if the max attempts is reached, otherwise it's nil
func (b *Backoff) Delay() (time.Duration, error) {
	if b.attempt == b.maxAttempts {
		return 0, &MaxAttemptsExceeded{}
	}
	b.attempt++
	return time.Duration(b.delay(b.attempt)), nil
}

func (b *Backoff) delay(n int) int64 {
	if n == 0 {
		return 0
	}
	pow := 1
	/*
		for {
			n--
			if n <= 1 {
				break
			}
			pow *= b.base
		}
	*/
	for ; n > 1; n-- {
		pow *= b.base
	}
	return b.scale * int64(pow)
}
