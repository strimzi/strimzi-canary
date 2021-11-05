//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

package services

import (
	"github.com/golang/glog"
	"github.com/hashicorp/go-uuid"
	"sync/atomic"
)

type Synchronizer struct {
	sync atomic.Value
}

func NewSynchronizer() *Synchronizer {
	var s Synchronizer
	s.sync.Store(getUuid())
	return &s
}

func (s *Synchronizer) Next() (next string) {
	next = getUuid()
	s.sync.Store(next)
	return
}

func (s *Synchronizer) Current() string {
	return s.sync.Load().(string)
}

func getUuid() (value string) {
	value, err := uuid.GenerateUUID()
	if err != nil {
		glog.Fatalf("Unexpected error generating UUID: %v", err)
	}
	return
}
