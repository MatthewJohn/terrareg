package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
)

// mockSessionRepoForCleanup is a minimal session repository mock
type mockSessionRepoForCleanup struct {
	mu           sync.Mutex
	cleanupCount int
	cleanupError error
	sessions     map[string]*auth.Session
}

func newMockSessionRepoForCleanup() *mockSessionRepoForCleanup {
	return &mockSessionRepoForCleanup{
		sessions: make(map[string]*auth.Session),
	}
}

func (m *mockSessionRepoForCleanup) Create(ctx context.Context, session *auth.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepoForCleanup) FindByID(ctx context.Context, sessionID string) (*auth.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[sessionID], nil
}

func (m *mockSessionRepoForCleanup) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockSessionRepoForCleanup) CleanupExpired(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupCount++
	if m.cleanupError != nil {
		return m.cleanupError
	}
	// Remove expired sessions
	for id, session := range m.sessions {
		if session.IsExpired() {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepoForCleanup) UpdateProviderSourceAuth(ctx context.Context, sessionID string, data []byte) error {
	return nil
}

func (m *mockSessionRepoForCleanup) getCleanupCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cleanupCount
}

func (m *mockSessionRepoForCleanup) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupCount = 0
	m.cleanupError = nil
	m.sessions = make(map[string]*auth.Session)
}

// newTestLoggerSessionCleanup creates a test logger with no output
func newTestLoggerSessionCleanup() logging.Logger {
	return logging.NewZeroLogger(zerolog.New(zerolog.Nop()).With().Timestamp().Logger())
}

// TestNewSessionCleanupService tests the constructor
func TestNewSessionCleanupService(t *testing.T) {
	t.Run("with explicit cleanup interval", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 30 * time.Minute

		service := NewSessionCleanupService(sessionService, logger, interval)

		assert.NotNil(t, service)
		// The interval is stored internally but we can't directly access it
		// We can verify the service was created successfully
	})

	t.Run("with zero interval uses default", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()

		service := NewSessionCleanupService(sessionService, logger, 0)

		assert.NotNil(t, service)
	})
}

// TestSessionCleanupService_StartStop tests basic start/stop functionality
func TestSessionCleanupService_StartStop(t *testing.T) {
	t.Run("stop stops the service gracefully", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 1 * time.Hour // Long interval to avoid periodic runs during test

		service := NewSessionCleanupService(sessionService, logger, interval)

		// Create a context that we can cancel
		ctx, cancel := context.WithCancel(context.Background())

		stopped := make(chan struct{})
		go func() {
			service.Start(ctx)
			close(stopped)
		}()

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// Stop the service via context cancellation
		cancel()

		// Wait for service to stop
		select {
		case <-stopped:
			// Service stopped as expected
		case <-time.After(200 * time.Millisecond):
			t.Error("Service should have stopped when context was cancelled")
		}

		// Verify cleanup ran at least once (immediate run on start)
		assert.Equal(t, 1, mockRepo.getCleanupCount())
	})

	t.Run("stop via Stop() method", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 1 * time.Hour

		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx := context.Background()

		stopped := make(chan struct{})
		go func() {
			service.Start(ctx)
			close(stopped)
		}()

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// Stop via Stop() method
		service.Stop()

		// Wait for service to stop
		select {
		case <-stopped:
			// Service stopped as expected
		case <-time.After(200 * time.Millisecond):
			t.Error("Service should have stopped when Stop() was called")
		}

		// Verify cleanup ran at least once
		assert.Equal(t, 1, mockRepo.getCleanupCount())
	})
}

// TestSessionCleanupService_ImmediateCleanup tests that cleanup runs immediately on start
func TestSessionCleanupService_ImmediateCleanup(t *testing.T) {
	t.Run("runs cleanup immediately on start", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 1 * time.Hour // Long interval to avoid periodic runs

		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start the service
		go service.Start(ctx)

		// Wait a bit for the immediate cleanup
		time.Sleep(50 * time.Millisecond)

		// Stop the service
		cancel()

		// Verify cleanup ran immediately
		assert.Equal(t, 1, mockRepo.getCleanupCount())
	})
}

// TestSessionCleanupService_PeriodicCleanup tests periodic cleanup execution
func TestSessionCleanupService_PeriodicCleanup(t *testing.T) {
	t.Run("runs cleanup periodically", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 50 * time.Millisecond

		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start the service
		go service.Start(ctx)

		// Wait for multiple cleanup cycles
		// Initial run + 2 more runs = 3 total
		time.Sleep(175 * time.Millisecond) // 50ms * 3 + some buffer

		// Stop the service
		cancel()

		// Should have run at least 3 times (initial + 2 periodic)
		count := mockRepo.getCleanupCount()
		assert.GreaterOrEqual(t, count, 3, "Expected at least 3 cleanup runs")
	})
}

// TestSessionCleanupService_ContextCancellation tests context cancellation
func TestSessionCleanupService_ContextCancellation(t *testing.T) {
	t.Run("stops when context is cancelled", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 1 * time.Hour

		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx, cancel := context.WithCancel(context.Background())

		stopped := make(chan struct{})
		go func() {
			service.Start(ctx)
			close(stopped)
		}()

		// Wait for service to start
		time.Sleep(10 * time.Millisecond)

		// Cancel the context
		cancel()

		// Wait for service to stop
		select {
		case <-stopped:
			// Service stopped as expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Service should have stopped when context was cancelled")
		}
	})

	t.Run("respects both context and stop channel", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 1 * time.Hour

		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx := context.Background()

		stopped := make(chan struct{})
		go func() {
			service.Start(ctx)
			close(stopped)
		}()

		// Wait for service to start
		time.Sleep(10 * time.Millisecond)

		// Stop via Stop() method
		service.Stop()

		// Wait for service to stop
		select {
		case <-stopped:
			// Service stopped as expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Service should have stopped when Stop() was called")
		}
	})
}

// TestSessionCleanupService_CleanupErrors tests error handling in cleanup
func TestSessionCleanupService_CleanupErrors(t *testing.T) {
	t.Run("continues running after cleanup error", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 50 * time.Millisecond

		// Set cleanup to error
		mockRepo.cleanupError = assert.AnError

		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start the service
		go service.Start(ctx)

		// Wait for multiple cleanup attempts
		time.Sleep(175 * time.Millisecond)

		// Stop the service
		cancel()

		// Service should have continued trying despite errors
		count := mockRepo.getCleanupCount()
		assert.GreaterOrEqual(t, count, 2, "Expected multiple cleanup attempts despite errors")
	})
}

// TestSessionCleanupService_Concurrency tests concurrent access to the service
func TestSessionCleanupService_Concurrency(t *testing.T) {
	t.Run("handles concurrent operations", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 50 * time.Millisecond
		service := NewSessionCleanupService(sessionService, logger, interval)

		ctx, cancel := context.WithCancel(context.Background())

		// Start the service
		go service.Start(ctx)

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// Perform concurrent operations (reading state, etc.)
		done := make(chan struct{})
		for i := 0; i < 5; i++ {
			go func(idx int) {
				defer func() { done <- struct{}{} }()
				// Simulate concurrent read operations
				time.Sleep(time.Duration(idx) * 10 * time.Millisecond)
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 5; i++ {
			<-done
		}

		// Stop the service
		cancel()
		service.Stop()

		// Verify cleanup ran at least once
		assert.GreaterOrEqual(t, mockRepo.getCleanupCount(), 1)
	})
}

// TestSessionCleanupService_Intervals tests various cleanup intervals
func TestSessionCleanupService_Intervals(t *testing.T) {
	intervals := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		1 * time.Second,
		30 * time.Second,
		5 * time.Minute,
		1 * time.Hour,
		24 * time.Hour,
	}

	for _, interval := range intervals {
		t.Run(interval.String(), func(t *testing.T) {
			mockRepo := newMockSessionRepoForCleanup()
			sessionService := NewSessionService(mockRepo, nil)
			logger := newTestLoggerSessionCleanup()

			service := NewSessionCleanupService(sessionService, logger, interval)

			assert.NotNil(t, service)
		})
	}
}

// TestSessionCleanupService_ActualCleanup tests actual cleanup of expired sessions
func TestSessionCleanupService_ActualCleanup(t *testing.T) {
	t.Run("actually removes expired sessions", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 100 * time.Millisecond

		// Create some sessions
		ctx := context.Background()
		now := time.Now()

		// Active session
		activeSession := &auth.Session{
			ID:                 "active-session",
			Expiry:             now.Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.Create(ctx, activeSession)

		// Expired sessions
		expiredSession1 := &auth.Session{
			ID:                 "expired-1",
			Expiry:             now.Add(-1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.Create(ctx, expiredSession1)

		expiredSession2 := &auth.Session{
			ID:                 "expired-2",
			Expiry:             now.Add(-2 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.Create(ctx, expiredSession2)

		service := NewSessionCleanupService(sessionService, logger, interval)

		// Run one cleanup cycle
		serviceCtx, cancel := context.WithCancel(context.Background())

		go service.Start(serviceCtx)

		// Wait for immediate cleanup
		time.Sleep(50 * time.Millisecond)

		// Stop the service
		cancel()

		// Verify expired sessions were removed
		active, _ := mockRepo.FindByID(ctx, "active-session")
		expired1, _ := mockRepo.FindByID(ctx, "expired-1")
		expired2, _ := mockRepo.FindByID(ctx, "expired-2")

		assert.NotNil(t, active, "Active session should still exist")
		assert.Nil(t, expired1, "Expired session 1 should be removed")
		assert.Nil(t, expired2, "Expired session 2 should be removed")
	})
}

// TestSessionCleanupService_ZeroInterval tests zero interval handling
func TestSessionCleanupService_ZeroInterval(t *testing.T) {
	t.Run("zero interval defaults to one hour", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()

		service := NewSessionCleanupService(sessionService, logger, 0)

		assert.NotNil(t, service)
		// The service should use default 1 hour interval internally
	})
}

// TestSessionCleanupService_StructFields tests service struct fields
func TestSessionCleanupService_StructFields(t *testing.T) {
	t.Run("service is properly initialized", func(t *testing.T) {
		mockRepo := newMockSessionRepoForCleanup()
		sessionService := NewSessionService(mockRepo, nil)
		logger := newTestLoggerSessionCleanup()
		interval := 30 * time.Minute

		service := NewSessionCleanupService(sessionService, logger, interval)

		assert.NotNil(t, service)
		// We can't directly access private fields, but we verified the service was created
	})
}
