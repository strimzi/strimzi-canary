//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package workers defines an interface for canary workers and related implementations
package workers

// Worker interface exposing main operations on canary workers
type Worker interface {
	Start()
	Stop()
	// Register metrics with prometheus default registery.
	// Decoupling metrics registration from Worker creation allows for multiple workers
	// to be created for testing without multiple registrations of the same metrics.
	RegisterMetrics()
}
