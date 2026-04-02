package models

import "time"

// Role définit les différents rôles utilisateurs
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleDirector   Role = "director"
	RoleAgent      Role = "agent"
	RoleClient     Role = "client"
)

// User représente un utilisateur de la plateforme
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone"`
	Role         Role      `json:"role"`
	AgencyID     *int64    `json:"agency_id,omitempty"`
	Avatar       string    `json:"avatar"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FullName retourne le nom complet de l'utilisateur
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsAdmin vérifie si l'utilisateur a des droits d'admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// CanManageAgency vérifie si l'utilisateur peut gérer une agence
func (u *User) CanManageAgency() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleDirector
}

// CanManageProperties vérifie si l'utilisateur peut gérer des biens
func (u *User) CanManageProperties() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleDirector || u.Role == RoleAgent
}
