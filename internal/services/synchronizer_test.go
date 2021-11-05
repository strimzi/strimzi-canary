package services

import (
	"testing"
)

func TestSynchronizer_IsInitialized(t *testing.T) {
	s := NewSynchronizer()
	if s.Current() == "" {
		t.Errorf("Unexpected empty value from new synchronizer")
	}
}

func TestSynchronizer_CurrentIdempotent(t *testing.T) {
	s := NewSynchronizer()
	current := s.Current()
	current1 := s.Current()
	if current != current1 {
		t.Errorf("Successive calls to Current should yield the same result")
	}
}

func TestSynchronizer_Next(t *testing.T) {
	s := NewSynchronizer()
	current := s.Current()
	s.Next()
	value := s.Current()

	if current == value {
		t.Errorf("Next expected to produce new value")
	}
}
