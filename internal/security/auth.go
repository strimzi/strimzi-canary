//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package security defining some security related tools
package security

import (
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

func SetAuthConfig(canaryConfig *config.CanaryConfig, saramaConfig *sarama.Config) error {

	if canaryConfig.SASLMechanism == sarama.SASLTypePlaintext {

		if canaryConfig.SASLUser == "" {
			return errors.New("SASL user must be specified")
		}
		if canaryConfig.SASLPassword == "" {
			return errors.New("SASL password must be specified")
		}
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.Version = sarama.SASLHandshakeV1
		saramaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(canaryConfig.SASLMechanism)
		saramaConfig.Net.SASL.User = canaryConfig.SASLUser
		saramaConfig.Net.SASL.Password = canaryConfig.SASLPassword
		return nil
	}
	return fmt.Errorf("SASL mechanism %s is not supported", canaryConfig.SASLMechanism)
}
