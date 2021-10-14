//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package util contains some utility functions
package util

import (
	"time"

	"github.com/golang/glog"
)

const (
	// max number of buckets in the time window ring buffer, which allows up to 1204 KB (8 byte * 128 buckets)
	maxBufferBuckets = 128
)

// ErrNoDataSamples defines the error can be raised when there are no data samples in the timw window ring
type ErrNoDataSamples struct{}

func (e *ErrNoDataSamples) Error() string {
	return "No data samples available in the time window ring"
}

// TimeWindowRing represents a struct leveraging a buffer of buckets to store values
// covering a sliding time window of specified "size" and sampled with "sampling" rate
// NOTE: internal buffer size is determined by time windows size and sampling rate capped at maxBufferBuckets
//
// tail (T): provides the first added value in the current sliding time window, it moves forward when the buffer is full and new values come
// head (H): provides the last added value in the current sliding time window, it always move forward since the beginning restarting when buffer is full
//
// --------------------------------------
// | v1 (T, H) |  |  | .... |  |  |  |  |  --> 1st value added
// --------------------------------------
//
// -----------------------------------------
// | v1 (T) | v2 (H) |  | .... |  |  |  |  |  --> 2nd value added
// -----------------------------------------
//
// ---------------------------------------------
// | v1 (T) | v2 | v3 | .... | vN (H) |  |  |  |  --> N values added in the time window buckets
// ---------------------------------------------
//
// -------------------------------------------------------
// | v1 (T) | v2 | v3 | .... | vN | vN+1 | vN+2 | vX (H) |  --> reached the end of the buffer
// -------------------------------------------------------
//
// ---------------------------------------------------------
// | vX+1 (H) | v2 (T) | v3 | .... | vN | vN+1 | vN+2 | vX |  --> start to fill the buffer using first localion kicking out old value (time window is moving)
// ---------------------------------------------------------
//
type TimeWindowRing struct {
	buffer   []uint64
	tail     int
	head     int
	count    int
	size     time.Duration
	sampling time.Duration
}

// NewTimeWindowRing returns an instance of TimeWindowRing
func NewTimeWindowRing(size time.Duration, sampling time.Duration) *TimeWindowRing {
	bufferSize := size / sampling
	if bufferSize > maxBufferBuckets {
		bufferSize = maxBufferBuckets
		glog.Warningf("Time window %d ms too wide with %d ms sampling; resized to %d ms",
			size, sampling, bufferSize*sampling)
	}
	return &TimeWindowRing{
		buffer:   make([]uint64, bufferSize),
		tail:     -1,
		head:     -1,
		size:     bufferSize * sampling,
		sampling: sampling,
	}
}

// Put allows to add a new sampled value in the time window ring buffer
func (rb *TimeWindowRing) Put(value uint64) {
	rb.head = (rb.head + 1) % len(rb.buffer)
	rb.buffer[rb.head] = value
	if rb.head == rb.tail {
		rb.tail = (rb.tail + 1) % len(rb.buffer)
	}
	if rb.tail == -1 {
		rb.tail = 0
	}
	rb.count++
}

// Head returns the sampled value at the head position
func (rb *TimeWindowRing) Head() uint64 {
	return rb.buffer[rb.head]
}

// Tail returns the sampled value at the tail position
func (rb *TimeWindowRing) Tail() uint64 {
	return rb.buffer[rb.tail]
}

// IsEmpty returns if there are no samples in the time window ring buffer yet
func (rb *TimeWindowRing) IsEmpty() bool {
	return rb.tail == -1 && rb.head == -1
}

// Count returns the number of samples in the time window ring buffer
func (rb *TimeWindowRing) Count() int {
	return rb.count
}
