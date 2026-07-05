package domain

type TeamRole string

const (
	RoleOwner  TeamRole = "owner"
	RoleAdmin  TeamRole = "admin"
	RoleMember TeamRole = "member"
)

func (t TeamRole) Validate() bool {
	return t == RoleOwner || t == RoleAdmin || t == RoleMember
}
