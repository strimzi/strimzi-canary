//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import "fmt"

// CanaryMessage defines the payload of a canary message
type CanaryMessage struct {
	ProducerID string `json:"producerId"`
	MessageID  int    `json:"messageId"`
	Timestamp  int64  `json:"timestamp"`
}

func (cm CanaryMessage) String() string {
	return fmt.Sprintf("{ProducerID:%s, MessageID:%d, Timestamp:%d}",
		cm.ProducerID, cm.MessageID, cm.Timestamp)
}
