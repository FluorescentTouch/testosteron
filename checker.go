package steron

import (
	"fmt"
	"sync/atomic"
	"testing"
)

type SequenceChecker struct {
	t       *testing.T
	pending atomic.Int32
	expects chan string
}

func NewSequenceChecker(t *testing.T) *SequenceChecker {
	return &SequenceChecker{
		t:       t,
		expects: make(chan string, 256),
	}
}

func (s *SequenceChecker) expect(value string) {
	s.pending.Add(1)
	select {
	case s.expects <- value:
		return
	default:
		s.t.Error("expect channel is full")
	}
}

func (s *SequenceChecker) Verify(got string) {
	s.pending.Add(-1)
	select {
	case expected := <-s.expects:
		if expected != got {
			s.t.Errorf("expected %q, got %q", expected, got)
		}
	default:
		s.t.Errorf("no expected value queued for %q", got)
	}
}

func (s *SequenceChecker) Remaining() int32 {
	return s.pending.Load()
}

func (s *SequenceChecker) AssertEmpty() error {
	if remaining := s.pending.Load(); remaining > 0 {
		return fmt.Errorf("unverified expectations: %d", remaining)
	}
	return nil
}
