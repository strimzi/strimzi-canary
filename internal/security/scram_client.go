//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// code inspired by the SCRAM client provided by the Sarama library at
// https://github.com/Shopify/sarama/blob/main/examples/sasl_scram_client/scram_client.go

package security

import (
	"crypto/sha256"
	"crypto/sha512"

	"github.com/xdg-go/scram"
)

var (
	SHA256 scram.HashGeneratorFcn = sha256.New
	SHA512 scram.HashGeneratorFcn = sha512.New
)

type CanarySCRAM struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (canarySCRAM *CanarySCRAM) Begin(username, password, authzID string) (err error) {
	if canarySCRAM.Client, err = canarySCRAM.HashGeneratorFcn.NewClient(username, password, authzID); err != nil {
		return err
	}
	canarySCRAM.ClientConversation = canarySCRAM.Client.NewConversation()
	return nil
}

func (canarySCRAM *CanarySCRAM) Step(challenge string) (response string, err error) {
	return canarySCRAM.ClientConversation.Step(challenge)
}
