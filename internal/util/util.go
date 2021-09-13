//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package util contains some utility functions
package util

import "time"

// NowInMilliseconds returns the current time in milliseconds
func NowInMilliseconds() int64 {
	return time.Now().UnixNano() / 1000000
}
