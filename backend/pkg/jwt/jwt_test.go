package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key-123"

	tests := []struct {
		name     string
		userID   string
		username string
		role     string
		wantErr  bool
	}{
		{
			name:     "Valid token generation",
			userID:   "user-123",
			username: "testuser",
			role:     "admin",
			wantErr:  false,
		},
		{
			name:     "Empty userID",
			userID:   "",
			username: "testuser",
			role:     "user",
			wantErr:  false, // Token generation allows empty values
		},
		{
			name:     "Empty username",
			userID:   "user-456",
			username: "",
			role:     "user",
			wantErr:  false,
		},
		{
			name:     "Empty role",
			userID:   "user-789",
			username: "testuser",
			role:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.username, tt.role, secret)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token format (should have 3 parts sep by .)
				assert.Contains(t, token, ".")
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	secret := "test-secret-key-123"

	tests := []struct {
		name         string
		setupToken   func() string
		secret       string
		wantUserID   string
		wantUsername string
		wantRole     string
		wantErr      bool
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := GenerateToken("user-123", "testuser", "admin", secret)
				return token
			},
			secret:       secret,
			wantUserID:   "user-123",
			wantUsername: "testuser",
			wantRole:     "admin",
			wantErr:      false,
		},
		{
			name: "Wrong secret",
			setupToken: func() string {
				token, _ := GenerateToken("user-123", "testuser", "admin", secret)
				return token
			},
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name: "Invalid token format",
			setupToken: func() string {
				return "invalid.token.format"
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Empty token",
			setupToken: func() string {
				return ""
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Malformed token",
			setupToken: func() string {
				return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.malformed"
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "Expired token",
			setupToken: func() string {
				claims := Claims{
					UserID:   "user-123",
					Username: "testuser",
					Role:     "admin",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // expired
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString
			},
			secret:  secret,
			wantErr: true, // Should fail due to expiration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupToken()
			claims, err := ParseToken(tokenString, tt.secret)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.wantUserID, claims.UserID)
				assert.Equal(t, tt.wantUsername, claims.Username)
				assert.Equal(t, tt.wantRole, claims.Role)
			}
		})
	}
}

// TestTokenRoundtrip 测试完整的token生成和解析流程
func TestTokenRoundtrip(t *testing.T) {
	secret := "test-secret-key-456"

	testCases := []struct {
		userID   string
		username string
		role     string
	}{
		{"user-001", "alice", "admin"},
		{"user-002", "bob", "user"},
		{"user-003", "charlie", "guest"},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			// Generate token
			token, err := GenerateToken(tc.userID, tc.username, tc.role, secret)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Parse token
			claims, err := ParseToken(token, secret)
			assert.NoError(t, err)
			assert.NotNil(t, claims)

			// Verify claims
			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.username, claims.Username)
			assert.Equal(t, tc.role, claims.Role)

			// Verify timestamps
			assert.True(t, claims.ExpiresAt.After(time.Now()))
			assert.True(t, claims.IssuedAt.Before(time.Now().Add(1*time.Second)))
		})
	}
}
