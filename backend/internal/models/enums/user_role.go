package enums

type UserRole string

const (
	UserRoleOwner  UserRole = "owner"
	UserRoleEditor UserRole = "editor"
	UserRoleViewer UserRole = "viewer"
)

func (r UserRole) IsValid() bool {
	return r == UserRoleOwner || r == UserRoleEditor || r == UserRoleViewer
}
