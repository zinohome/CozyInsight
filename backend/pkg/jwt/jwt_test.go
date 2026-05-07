package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_GenerateAndParse(t *testing.T) {
	m := NewManager("test-secret", 2*time.Hour)

	token, err := m.Generate(1, "admin", true)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := m.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.True(t, claims.IsAdmin)
	assert.Equal(t, "1", claims.Subject)
}

func TestManager_Parse_Expired(t *testing.T) {
	m := NewManager("test-secret", -1*time.Hour)

	token, err := m.Generate(1, "admin", false)
	require.NoError(t, err)

	_, err = m.Parse(token)
	assert.Error(t, err)
}

func TestManager_Parse_InvalidSecret(t *testing.T) {
	m1 := NewManager("secret-a", 2*time.Hour)
	m2 := NewManager("secret-b", 2*time.Hour)

	token, err := m1.Generate(1, "admin", false)
	require.NoError(t, err)

	_, err = m2.Parse(token)
	assert.Error(t, err)
}

func TestManager_Parse_Empty(t *testing.T) {
	m := NewManager("test-secret", 2*time.Hour)
	_, err := m.Parse("")
	assert.Error(t, err)
}
