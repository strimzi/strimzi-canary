//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import (
	"time"
)

const (
	ScaleDefault       = 200 * time.Millisecond
	MaxDefault         = 5 * time.Minute
	MaxAttemptsDefault = 6
)

// Backoff encapsulates computing delays for an exponential back-off, when an operation has to be retried
type Backoff struct {
	maxAttempts int
	scale       time.Duration
	max         time.Duration
	attempt     int
}

// MaxAttemptsExceeded defines the error for the max attempts exceeded
type MaxAttemptsExceeded struct{}

// Overflow on computed delay
type BackoffDelayOverflow struct{}

func (e *MaxAttemptsExceeded) Error() string {
	return "Maximum number of attempts exceeded"
}

func (e *BackoffDelayOverflow) Error() string {
	return "Overflow on the computed backoff delay"
}

// NewBackoff returns an instance of a Backoff struct
func NewBackoff(maxAttempts int, scale time.Duration, max time.Duration) *Backoff {
	actualScale := scale
	if actualScale <= 0 {
		actualScale = ScaleDefault
	}
	actualMax := max
	if actualMax <= 0 {
		actualMax = MaxDefault
	}
	backoff := Backoff{
		maxAttempts: maxAttempts,
		scale:       actualScale,
		max:         actualMax,
		attempt:     0,
	}
	return &backoff
}

// Delay computes a delay in terms of Duration (nanoseconds) based on the current Backoff instance configuration
// Returns the delay in terms of Duration (nanoseconds) and an error if the max attempts is reached, otherwise it's nil
func (b *Backoff) Delay() (time.Duration, error) {
	if b.attempt == b.maxAttempts {
		return 0, &MaxAttemptsExceeded{}
	}
	delay := time.Duration(b.scale * 1 << b.attempt)
	if delay < 0 {
		return 0, &BackoffDelayOverflow{}
	}
	if delay > b.max {
		delay = b.max
	}
	b.attempt++
	return delay, nil
}
