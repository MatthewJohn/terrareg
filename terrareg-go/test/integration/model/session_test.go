package model

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// generateSessionID generates a random session ID similar to Python's secrets.token_urlsafe
func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// TestSession_Create tests creating sessions
func TestSession_Create(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	sessionID := generateSessionID()
	expiry := time.Now().Add(60 * time.Minute)

	session := sqldb.SessionDB{
		ID:     sessionID,
		Expiry: expiry,
	}

	err := db.DB.Create(&session).Error
	require.NoError(t, err)

	assert.Equal(t, sessionID, session.ID)
	assert.WithinDuration(t, expiry, session.Expiry, time.Second)
}

// TestSession_Delete tests deleting sessions
func TestSession_Delete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	sessionID := generateSessionID()
	expiry := time.Now().Add(1 * time.Minute)

	session := sqldb.SessionDB{
		ID:     sessionID,
		Expiry: expiry,
	}

	err := db.DB.Create(&session).Error
	require.NoError(t, err)

	// Verify session exists
	var retrieved sqldb.SessionDB
	err = db.DB.Where("id = ?", sessionID).First(&retrieved).Error
	require.NoError(t, err)

	// Delete session
	err = db.DB.Delete(&session).Error
	require.NoError(t, err)

	// Verify deleted
	err = db.DB.Where("id = ?", sessionID).First(&retrieved).Error
	assert.Error(t, err)
}

// TestSession_CheckValidSession tests checking valid sessions
func TestSession_CheckValidSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	sessionID := generateSessionID()
	expiry := time.Now().Add(1 * time.Minute)

	session := sqldb.SessionDB{
		ID:     sessionID,
		Expiry: expiry,
	}

	err := db.DB.Create(&session).Error
	require.NoError(t, err)

	// Check session exists
	var retrieved sqldb.SessionDB
	err = db.DB.Where("id = ?", sessionID).First(&retrieved).Error
	require.NoError(t, err)
	assert.Equal(t, sessionID, retrieved.ID)
}

// TestSession_CheckExpiredSession tests checking expired sessions
func TestSession_CheckExpiredSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	sessionID := generateSessionID()
	expiry := time.Now().Add(-1 * time.Minute) // Expired

	session := sqldb.SessionDB{
		ID:     sessionID,
		Expiry: expiry,
	}

	err := db.DB.Create(&session).Error
	require.NoError(t, err)

	// Session exists in database
	var retrieved sqldb.SessionDB
	err = db.DB.Where("id = ?", sessionID).First(&retrieved).Error
	require.NoError(t, err)

	// Verify it's expired
	assert.True(t, retrieved.Expiry.Before(time.Now()))
}

// TestSession_CheckNonExistentSession tests checking non-existent sessions
func TestSession_CheckNonExistentSession(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	// Check non-existent session
	var retrieved sqldb.SessionDB
	err := db.DB.Where("id = ?", "nonexistent-session-id").First(&retrieved).Error
	assert.Error(t, err)
}

// TestSession_CleanupExpiredSessions tests cleaning up expired sessions
func TestSession_CleanupExpiredSessions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	// Create expired and non-expired sessions
	expiredSessionID := generateSessionID()
	validSessionID := generateSessionID()

	expiredSession := sqldb.SessionDB{
		ID:     expiredSessionID,
		Expiry: time.Now().Add(-10 * time.Minute),
	}

	validSession := sqldb.SessionDB{
		ID:     validSessionID,
		Expiry: time.Now().Add(10 * time.Minute),
	}

	err := db.DB.Create(&expiredSession).Error
	require.NoError(t, err)
	err = db.DB.Create(&validSession).Error
	require.NoError(t, err)

	// Delete expired sessions (where expiry < now)
	err = db.DB.Where("expiry < ?", time.Now()).Delete(&sqldb.SessionDB{}).Error
	require.NoError(t, err)

	// Verify only valid session remains
	var count int64
	err = db.DB.Model(&sqldb.SessionDB{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	var retrieved sqldb.SessionDB
	err = db.DB.Where("id = ?", validSessionID).First(&retrieved).Error
	require.NoError(t, err)
	assert.Equal(t, validSessionID, retrieved.ID)
}

// TestSession_InvalidSessionIDs tests handling invalid session IDs
func TestSession_InvalidSessionIDs(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	testCases := []struct {
		name      string
		sessionID string
		wantError bool
	}{
		{"empty string", "", true},
		{"non-existent", "NMBfiFLW3EQjHVFZlM5T6Tomzcj3fEGe87Hc1u38afA", true},
		{"invalid characters", "`\"'@$#}+=", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var retrieved sqldb.SessionDB
			err := db.DB.Where("id = ?", tc.sessionID).First(&retrieved).Error
			if tc.wantError {
				assert.Error(t, err)
			}
		})
	}
}

// TestSession_ProviderSourceAuth tests storing and retrieving provider source auth
func TestSession_ProviderSourceAuth(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	sessionID := generateSessionID()
	authData := []byte(`{"token": "test-token", "expiry": "2024-01-01"}`)

	session := sqldb.SessionDB{
		ID:                 sessionID,
		Expiry:             time.Now().Add(60 * time.Minute),
		ProviderSourceAuth: authData,
	}

	err := db.DB.Create(&session).Error
	require.NoError(t, err)

	// Retrieve and verify
	var retrieved sqldb.SessionDB
	err = db.DB.Where("id = ?", sessionID).First(&retrieved).Error
	require.NoError(t, err)
	assert.Equal(t, authData, retrieved.ProviderSourceAuth)
}

// TestSession_ExpiryTimeRange tests that expiry times are within expected ranges
func TestSession_ExpiryTimeRange(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean session table
	db.DB.Exec("DELETE FROM session")

	sessionID := generateSessionID()
	expectedExpiry := time.Now().Add(60 * time.Minute)

	session := sqldb.SessionDB{
		ID:     sessionID,
		Expiry: expectedExpiry,
	}

	err := db.DB.Create(&session).Error
	require.NoError(t, err)

	// Verify expiry is within 2 minutes of expected
	var retrieved sqldb.SessionDB
	err = db.DB.Where("id = ?", sessionID).First(&retrieved).Error
	require.NoError(t, err)

	// Check expiry is about the correct time (within 2 minutes)
	assert.True(t, retrieved.Expiry.After(time.Now().Add(58*time.Minute)))
	assert.True(t, retrieved.Expiry.Before(time.Now().Add(62*time.Minute)))
}
