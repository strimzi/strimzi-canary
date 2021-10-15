//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package util contains some utility functions
package util

import (
	"testing"
)

func TestRing(t *testing.T) {
	ring := NewTimeWindowRing(10000, 2000)
	ring.Put(1)
	ring.Put(2)
	ring.Put(3)
	ring.Put(4)
	ring.Put(5)
	if ring.Tail() != 1 {
		t.Errorf("got = %d, want = %d", ring.Tail(), 1)
	}
	if ring.Head() != 5 {
		t.Errorf("got = %d, want = %d", ring.Head(), 5)
	}
}

func TestRingRolling(t *testing.T) {
	ring := NewTimeWindowRing(10000, 2000)
	ring.Put(1)
	ring.Put(2)
	ring.Put(3)
	ring.Put(4)
	ring.Put(5)
	ring.Put(6)
	if ring.Tail() != 2 {
		t.Errorf("got = %d, want = %d", ring.Tail(), 2)
	}
	if ring.Head() != 6 {
		t.Errorf("got = %d, want = %d", ring.Head(), 6)
	}
}
