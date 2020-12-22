//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import (
	"testing"
	"time"
)

func TestBackoffDelayDefault(t *testing.T) {
	b := NewBackoff(MaxAttemptsDefault, ScaleDefault)
	want := time.Duration(200)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}

func TestBackoffMoreDelayDefault(t *testing.T) {
	b := NewBackoff(MaxAttemptsDefault, ScaleDefault)
	want := time.Duration(200)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = time.Duration(400)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = time.Duration(800)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}

func TestBackoffMaxAttemptsExceeded(t *testing.T) {
	maxAttempts := 3
	b := NewBackoff(maxAttempts, ScaleDefault)

	for i := 0; i < maxAttempts; i++ {
		if _, err := b.Delay(); err != nil {
			t.Errorf("Error should be nil, got = %v", err)
			return
		}
	}
	if _, err := b.Delay(); err == nil {
		t.Errorf("Expecting max attempts exceeded error")
	}
}

func TestBackoffMoreDelay(t *testing.T) {
	b := NewBackoff(3, 5000)
	want := time.Duration(5000)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = time.Duration(10000)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = time.Duration(20000)
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}
