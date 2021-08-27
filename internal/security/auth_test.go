//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// +build unit_test

// Package security defining some security related tools
package security

import (
	"os"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

func TestNoAuth(t *testing.T) {
	canaryConfig := config.NewCanaryConfig()
	saramaConfig := sarama.NewConfig()
	e := SetAuthConfig(canaryConfig, saramaConfig)
	if e == nil {
		t.Fail()
	}
}

func TestSASLPlainAuthNoUserPassword(t *testing.T) {
	os.Setenv(config.SASLMechanismEnvVar, "PLAIN")
	canaryConfig := config.NewCanaryConfig()
	saramaConfig := sarama.NewConfig()
	e := SetAuthConfig(canaryConfig, saramaConfig)
	if e == nil {
		t.Fail()
	}
}

func TestSASLPlainAuth(t *testing.T) {
	os.Setenv(config.SASLMechanismEnvVar, "PLAIN")
	os.Setenv(config.SASLUserEnvVar, "user")
	os.Setenv(config.SASLPasswordEnvVar, "password")
	canaryConfig := config.NewCanaryConfig()
	saramaConfig := sarama.NewConfig()
	e := SetAuthConfig(canaryConfig, saramaConfig)
	if e != nil ||
		!saramaConfig.Net.SASL.Enable || saramaConfig.Net.SASL.Version != sarama.SASLHandshakeV1 ||
		saramaConfig.Net.SASL.Mechanism != sarama.SASLMechanism(canaryConfig.SASLMechanism) ||
		saramaConfig.Net.SASL.User != canaryConfig.SASLUser || saramaConfig.Net.SASL.Password != canaryConfig.SASLPassword {
		t.Fail()
	}
}
