package auth

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleAuditor Role = "auditor"
	RoleUser    Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleAuditor, RoleUser:
		return true
	default:
		return false
	}
}
