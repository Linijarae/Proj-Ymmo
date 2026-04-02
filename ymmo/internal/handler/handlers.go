package handler

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"ymmo/internal/middleware"
	"ymmo/internal/models"
	"ymmo/internal/service"
)

// TemplateData contient les données passées à tous les templates
type TemplateData struct {
	Title      string
	User       *models.User
	Claims     *service.Claims
	Flash      string
	FlashType  string // "success", "error", "info"
	Data       interface{}
	Year       int
}

// Handlers regroupe tous les handlers de l'application
type Handlers struct {
	auth     *service.AuthService
	property *service.PropertyService
	users    interface{ FindByID(int64) (*models.User, error) }
	tmpl     *template.Template
	agencies interface {
		ListAll() ([]*models.Agency, error)
		FindByID(int64) (*models.Agency, error)
		Create(*models.Agency) error
		Update(*models.Agency) error
		Delete(int64) error
	}
	contacts interface {
		ListAll() ([]*models.ContactRequest, error)
		UpdateStatus(int64, string) error
	}
	usersRepo interface {
		FindByID(int64) (*models.User, error)
		ListAll(string) ([]*models.User, error)
		SetRole(int64, models.Role, *int64) error
		SetActive(int64, bool) error
		Update(*models.User) error
	}
}

// NewHandlers crée les handlers en chargeant les templates
func NewHandlers(auth *service.AuthService, property *service.PropertyService, usersRepo interface {
	FindByID(int64) (*models.User, error)
	ListAll(string) ([]*models.User, error)
	SetRole(int64, models.Role, *int64) error
	SetActive(int64, bool) error
	Update(*models.User) error
}, agenciesRepo interface {
	ListAll() ([]*models.Agency, error)
	FindByID(int64) (*models.Agency, error)
	Create(*models.Agency) error
	Update(*models.Agency) error
	Delete(int64) error
}, contactsRepo interface {
	ListAll() ([]*models.ContactRequest, error)
	UpdateStatus(int64, string) error
}) *Handlers {

	tmpl := loadTemplates()
	return &Handlers{
		auth:      auth,
		property:  property,
		usersRepo: usersRepo,
		agencies:  agenciesRepo,
		contacts:  contactsRepo,
		tmpl:      tmpl,
	}
}

// loadTemplates charge tous les templates HTML
func loadTemplates() *template.Template {
	funcMap := template.FuncMap{
		"formatPrice": func(price float64) string {
			if price >= 1000000 {
				return fmt.Sprintf("%.1f M€", price/1000000)
			} else if price >= 1000 {
				return fmt.Sprintf("%s €", formatNumber(int(price)))
			}
			return fmt.Sprintf("%d €/mois", int(price))
		},
		"formatPriceRaw": func(price float64) string {
			return fmt.Sprintf("%s €", formatNumber(int(price)))
		},
		"formatSurface": func(s float64) string {
			return fmt.Sprintf("%.0f m²", s)
		},
		"toInt": func(f float64) int {
			return int(f)
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"ceil": func(f float64) int { return int(math.Ceil(f)) },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"hasRole": func(claims *service.Claims, role string) bool {
			if claims == nil {
				return false
			}
			return string(claims.Role) == role
		},
		"isAdmin": func(claims *service.Claims) bool {
			if claims == nil {
				return false
			}
			return claims.Role == models.RoleSuperAdmin
		},
		"canManage": func(claims *service.Claims) bool {
			if claims == nil {
				return false
			}
			return claims.Role == models.RoleSuperAdmin || claims.Role == models.RoleDirector || claims.Role == models.RoleAgent
		},
		"statusClass": func(status interface{}) string {
			s := fmt.Sprint(status)
			switch s {
			case "for_sale":
				return "badge-sale"
			case "for_rent":
				return "badge-rent"
			case "sold":
				return "badge-sold"
			case "rented":
				return "badge-rented"
			default:
				return ""
			}
		},
		"statusLabel": func(status interface{}) string {
			s := fmt.Sprint(status)
			switch s {
			case "for_sale":
				return "À vendre"
			case "for_rent":
				return "À louer"
			case "sold":
				return "Vendu"
			case "rented":
				return "Loué"
			default:
				return s
			}
		},
		"subTypeLabel": func(t interface{}) string {
			labels := map[string]string{
				"apartment": "Appartement", "house": "Maison", "villa": "Villa",
				"studio": "Studio", "loft": "Loft", "duplex": "Duplex",
				"terraced": "Maison de ville", "office": "Bureau", "retail": "Commerce",
				"warehouse": "Entrepôt", "building": "Immeuble", "land": "Terrain",
			}
			s := fmt.Sprint(t)
			if l, ok := labels[s]; ok {
				return l
			}
			return s
		},
		"toString": fmt.Sprint,
		"divF": func(a, b float64) string {
			if b == 0 {
				return "0"
			}
			return fmt.Sprintf("%.0f", a/b)
		},
		"currentYear": func() int { return time.Now().Year() },
		"formatDate": func(t time.Time) string {
			return t.Format("02/01/2006")
		},
		"truncate": func(s string, n int) string {
			if len([]rune(s)) <= n {
				return s
			}
			return string([]rune(s)[:n]) + "…"
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"initial": func(s string) string {
			if len(s) == 0 {
				return "?"
			}
			r := []rune(s)
			return strings.ToUpper(string(r[0:1]))
		},
		"roleLabel": func(role string) string {
			labels := map[string]string{
				"super_admin": "Super Admin",
				"director":    "Directeur",
				"agent":       "Agent",
				"client":      "Client",
			}
			if l, ok := labels[role]; ok {
				return l
			}
			return role
		},
		"contactStatusLabel": func(status string) string {
			labels := map[string]string{
				"pending":   "En attente",
				"processed": "Traité",
				"closed":    "Fermé",
			}
			if l, ok := labels[status]; ok {
				return l
			}
			return status
		},
	}

	// Chemin vers les templates
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..", "web", "templates")

	pattern := filepath.Join(root, "*.html")
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		log.Fatalf("Erreur de chargement des templates: %v", err)
	}
	return tmpl
}

// render exécute un template avec les données
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	claims := middleware.GetClaims(r)

	var currentUser *models.User
	if claims != nil {
		u, err := h.usersRepo.FindByID(claims.UserID)
		if err == nil {
			currentUser = u
		}
	}

	td := &TemplateData{
		Title:   "Ymmo - Votre expert immobilier",
		User:    currentUser,
		Claims:  claims,
		Data:    data,
		Year:    time.Now().Year(),
	}

	// Flash message depuis cookie
	if cookie, err := r.Cookie("flash"); err == nil {
		parts := strings.SplitN(cookie.Value, ":", 2)
		if len(parts) == 2 {
			td.FlashType = parts[0]
			td.Flash = parts[1]
		}
		// Supprimer le cookie flash
		http.SetCookie(w, &http.Cookie{
			Name:   "flash",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, td); err != nil {
		log.Printf("Erreur de rendu template %s: %v", name, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// setFlash définit un message flash dans un cookie
func setFlash(w http.ResponseWriter, flashType, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    flashType + ":" + msg,
		Path:     "/",
		MaxAge:   60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// redirect redirige avec un message flash
func redirectWithFlash(w http.ResponseWriter, r *http.Request, url, flashType, msg string) {
	setFlash(w, flashType, msg)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// formatNumber formate un nombre avec des espaces
func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += " "
		}
		result += string(c)
	}
	return result
}
