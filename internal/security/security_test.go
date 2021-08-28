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

	"github.com/strimzi/strimzi-canary/internal/config"
)

func TestSystemCertsPool(t *testing.T) {
	os.Setenv(config.TLSEnabledEnvVar, "true")
	canaryConfig := config.NewCanaryConfig()
	_, e := NewTLSConfig(canaryConfig)
	if e != nil {
		t.Fail()
	}
}
