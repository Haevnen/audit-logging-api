package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Haevnen/audit-logging-api/internal/auth"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		role     auth.Role
		expected bool
	}{
		{"Valid Admin", auth.RoleAdmin, true},
		{"Valid Auditor", auth.RoleAuditor, true},
		{"Valid User", auth.RoleUser, true},
		{"Invalid Empty", auth.Role(""), false},
		{"Invalid Random", auth.Role("superman"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.IsValid())
		})
	}
}
