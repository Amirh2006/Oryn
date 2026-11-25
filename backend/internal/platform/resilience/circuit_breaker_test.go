package resilience

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_TripsAndRecovers(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures:       2,
		ResetTimeout:      50 * time.Millisecond,
		HalfOpenSuccesses: 1,
	})

	assert.Equal(t, StateClosed, cb.State())

	// 1st failure
	err := cb.Execute(func() error {
		return errors.New("timeout")
	})
	assert.Error(t, err)
	assert.Equal(t, StateClosed, cb.State())

	// 2nd failure -> should trip circuit to OPEN
	err = cb.Execute(func() error {
		return errors.New("timeout")
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())

	// Next call should fail immediately without executing operation
	executed := false
	err = cb.Execute(func() error {
		executed = true
		return nil
	})
	assert.ErrorIs(t, err, ErrCircuitOpen)
	assert.False(t, executed)

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// In half-open state, success should close the circuit
	err = cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}
