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
	b := NewBackoff(MaxAttemptsDefault, ScaleDefault, MaxDefault)
	want := 200 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}

func TestBackoffMax(t *testing.T) {
	b := NewBackoff(MaxAttemptsDefault, 6*time.Minute, MaxDefault)
	want := 5 * time.Minute
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}

func TestBackoffMoreDelayDefault(t *testing.T) {
	b := NewBackoff(MaxAttemptsDefault, ScaleDefault, MaxDefault)
	want := 200 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = 400 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = 800 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}

func TestBackoffMaxAttemptsExceeded(t *testing.T) {
	maxAttempts := 3
	b := NewBackoff(maxAttempts, ScaleDefault, MaxDefault)

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
	b := NewBackoff(3, 5000*time.Millisecond, MaxDefault)
	want := 5000 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = 10000 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
	want = 20000 * time.Millisecond
	if delay, _ := b.Delay(); delay != want {
		t.Errorf("Delay: got = %v, want = %v", delay, want)
	}
}
