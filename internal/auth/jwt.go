package auth

//go:generate mockgen -source=jwt.go -destination=mocks/jwt.go

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ManagerInterface interface {
	GenerateToken(userID, tenantID string, role Role, ttl time.Duration) (string, error)
	ParseToken(tokenStr string) (*Claims, error)
}

type Manager struct {
	secretKey []byte
}

type Claims struct {
	UserID   string
	TenantID string
	Role     Role
	jwt.RegisteredClaims
}

// NewManager creates a new JWT manager with secret from config
func NewManager(secret string) *Manager {
	return &Manager{secretKey: []byte(secret)}
}

func (m *Manager) GenerateToken(userID, tenantID string, role Role, ttl time.Duration) (string, error) {
	if !role.IsValid() {
		return "", jwt.ErrInvalidKeyType
	}

	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *Manager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, jwt.ErrTokenExpired
	}

	return claims, nil
}
