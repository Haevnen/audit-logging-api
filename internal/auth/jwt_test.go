package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Haevnen/audit-logging-api/internal/auth"
)

func TestGenerateAndParseToken_Success(t *testing.T) {
	manager := auth.NewManager("secret-key")

	// generate token
	tokenStr, err := manager.GenerateToken("u1", "t1", auth.RoleUser, time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// parse token
	claims, err := manager.ParseToken(tokenStr)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "u1", claims.UserID)
	assert.Equal(t, "t1", claims.TenantID)
	assert.Equal(t, auth.RoleUser, claims.Role)
}

func TestGenerateToken_InvalidRole(t *testing.T) {
	manager := auth.NewManager("secret")

	// use invalid role (empty string or something not in Role enum)
	tokenStr, err := manager.GenerateToken("u1", "t1", auth.Role("invalid"), time.Minute)
	assert.Error(t, err)
	assert.Empty(t, tokenStr)
}

func TestParseToken_Expired(t *testing.T) {
	manager := auth.NewManager("secret-key")

	// generate expired token (ttl = -1s)
	tokenStr, err := manager.GenerateToken("u1", "t1", auth.RoleUser, -1*time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// parse should fail
	claims, err := manager.ParseToken(tokenStr)
	assert.Error(t, err)
	assert.Nil(t, claims)

	// ✅ use errors.Is instead of direct equality
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired), "expected jwt.ErrTokenExpired, got %v", err)
}

func TestParseToken_Tampered(t *testing.T) {
	manager1 := auth.NewManager("secret1")
	manager2 := auth.NewManager("secret2")

	// token signed with secret1
	tokenStr, err := manager1.GenerateToken("u1", "t1", auth.RoleUser, time.Minute)
	assert.NoError(t, err)

	// parse with manager2 (wrong secret) → should fail
	claims, err := manager2.ParseToken(tokenStr)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
