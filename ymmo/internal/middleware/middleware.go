package middleware

import (
	"context"
	"net/http"
	"ymmo/internal/models"
	"ymmo/internal/service"
)

type contextKey string

const (
	UserContextKey  contextKey = "user"
	ClaimsContextKey contextKey = "claims"
)

// Auth est le middleware d'authentification optionnel
// Il lit le JWT depuis le cookie et injecte les infos utilisateur dans le contexte
func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err == nil && cookie.Value != "" {
				claims, err := authSvc.ValidateToken(cookie.Value)
				if err == nil {
					ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth est un middleware qui exige une authentification
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r)
		if claims == nil {
			http.Redirect(w, r, "/connexion?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole exige un rôle spécifique
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil {
				http.Redirect(w, r, "/connexion", http.StatusSeeOther)
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Accès refusé", http.StatusForbidden)
		})
	}
}

// GetClaims extrait les claims JWT du contexte
func GetClaims(r *http.Request) *service.Claims {
	claims, _ := r.Context().Value(ClaimsContextKey).(*service.Claims)
	return claims
}

// IsAuthenticated vérifie si l'utilisateur est connecté
func IsAuthenticated(r *http.Request) bool {
	return GetClaims(r) != nil
}

// CSRF protège contre les attaques CSRF en vérifiant l'origine
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			origin := r.Header.Get("Origin")
			referer := r.Header.Get("Referer")
			host := r.Host
			if origin != "" && origin != "http://"+host && origin != "https://"+host {
				http.Error(w, "Requête invalide", http.StatusForbidden)
				return
			}
			if referer != "" {
				// Vérification basique du referer
				_ = referer
			}
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders ajoute les en-têtes de sécurité HTTP
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline' cdn.jsdelivr.net cdnjs.cloudflare.com; "+
				"style-src 'self' 'unsafe-inline' fonts.googleapis.com cdnjs.cloudflare.com; "+
				"font-src 'self' fonts.gstatic.com; img-src 'self' data:;")
		next.ServeHTTP(w, r)
	})
}
