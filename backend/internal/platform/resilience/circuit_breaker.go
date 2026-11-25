package resilience

import (
	"sync"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF-OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

var (
	ErrCircuitOpen = domain.NewAppError(domain.ErrCodeExternalProvider, "circuit breaker is open, upstream service unavailable", nil)
)

// CircuitBreakerConfig defines circuit breaker settings
type CircuitBreakerConfig struct {
	MaxFailures      int           // Number of consecutive failures before opening
	ResetTimeout     time.Duration // Time to wait before entering half-open state
	HalfOpenSuccesses int          // Required consecutive successes in half-open state before closing
}

// CircuitBreaker manages fault tolerance and fail-fast behavior for external services
type CircuitBreaker struct {
	mu                sync.RWMutex
	state             State
	failureCount      int
	successCount      int
	lastFailureTime   time.Time
	maxFailures       int
	resetTimeout      time.Duration
	halfOpenSuccesses int
}

// NewCircuitBreaker creates a new CircuitBreaker
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 5
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = 30 * time.Second
	}
	if cfg.HalfOpenSuccesses <= 0 {
		cfg.HalfOpenSuccesses = 2
	}

	return &CircuitBreaker{
		state:             StateClosed,
		maxFailures:       cfg.MaxFailures,
		resetTimeout:      cfg.ResetTimeout,
		halfOpenSuccesses: cfg.HalfOpenSuccesses,
	}
}

// State returns the current circuit state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		return StateHalfOpen
	}
	return cb.state
}

// Execute wraps an operation in circuit breaker protection
func (cb *CircuitBreaker) Execute(op func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := op()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		if cb.state == StateHalfOpen || cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Success path
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenSuccesses {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}

	return nil
}
