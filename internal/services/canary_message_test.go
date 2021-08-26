//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import (
	"testing"

	"github.com/Shopify/sarama"
)

func TestCanaryMessage(t *testing.T) {
	cm := CanaryMessage{
		ProducerID: "producer-id",
		MessageID:  0,
		Timestamp:  12345,
	}
	want := "{\"producerId\":\"producer-id\",\"messageId\":0,\"timestamp\":12345}"
	if cm.Json() != want {
		t.Errorf("JSON got = %s, want = %s", cm.Json(), want)
	}
}

func TestCanaryMessageEncode(t *testing.T) {
	cm := CanaryMessage{
		ProducerID: "producer-id",
		MessageID:  0,
		Timestamp:  12345,
	}
	encoder := sarama.StringEncoder(cm.Json())
	bytes, _ := encoder.Encode()

	decodedCm := NewCanaryMessage(bytes)

	if cm != decodedCm {
		t.Errorf("got = %v, want = %v", decodedCm, cm)
	}

	wrong := "{\"producerId\":\"producer-id\",\"messageId\":1,\"timestamp\":67890}"

	encoder = sarama.StringEncoder(wrong)
	bytes, _ = encoder.Encode()

	decodedCm = NewCanaryMessage(bytes)

	if cm == decodedCm {
		t.Errorf("got %v should be different from %v", decodedCm, cm)
	}
}
