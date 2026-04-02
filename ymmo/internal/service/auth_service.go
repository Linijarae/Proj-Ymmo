package service

import (
	"errors"
	"time"
	"ymmo/internal/models"
	"ymmo/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials est retourné lors d'une tentative de connexion échouée
var ErrInvalidCredentials = errors.New("email ou mot de passe incorrect")

// ErrEmailExists est retourné si l'email est déjà utilisé
var ErrEmailExists = errors.New("cet email est déjà utilisé")

// ErrAccountDisabled est retourné si le compte est désactivé
var ErrAccountDisabled = errors.New("ce compte a été désactivé")

// Claims contient les informations du JWT
type Claims struct {
	UserID   int64       `json:"user_id"`
	Email    string      `json:"email"`
	Role     models.Role `json:"role"`
	AgencyID *int64      `json:"agency_id,omitempty"`
	jwt.RegisteredClaims
}

// AuthService gère l'authentification
type AuthService struct {
	users     *repository.UserRepository
	jwtSecret []byte
}

// NewAuthService crée un nouveau service d'authentification
func NewAuthService(users *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register crée un nouveau compte client
func (s *AuthService) Register(email, password, firstName, lastName, phone string) (*models.User, error) {
	existing, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
		Role:         models.RoleClient,
		IsActive:     true,
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login authentifie un utilisateur et retourne un token JWT
func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, ErrInvalidCredentials
	}
	if !u.IsActive {
		return "", nil, ErrAccountDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// ValidateToken valide un token JWT et retourne les claims
func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("méthode de signature invalide")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("token invalide")
	}
	return claims, nil
}

// ChangePassword change le mot de passe d'un utilisateur
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("utilisateur introuvable")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("ancien mot de passe incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(userID, string(hash))
}

// HashPassword génère un hash bcrypt (utilitaire)
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func (s *AuthService) generateToken(u *models.User) (string, error) {
	claims := &Claims{
		UserID:   u.ID,
		Email:    u.Email,
		Role:     u.Role,
		AgencyID: u.AgencyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
