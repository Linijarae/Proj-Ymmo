package handler

import (
	"errors"
	"net/http"
	"ymmo/internal/middleware"
	"ymmo/internal/service"
)

// HomeHandler affiche la page d'accueil
func (h *Handlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	featured, err := h.property.ListFeatured(6)
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	agencies, _ := h.agencies.ListAll()
	summary, _ := h.property.GetAnalyticsSummary()
	cityStats, _ := h.property.GetCityStats(6)

	h.render(w, r, "index.html", map[string]interface{}{
		"FeaturedProperties": featured,
		"Agencies":           agencies,
		"Summary":            summary,
		"CityStats":          cityStats,
	})
}

// LoginHandler affiche le formulaire de connexion
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/tableau-de-bord", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		h.render(w, r, "login.html", map[string]interface{}{
			"Redirect": r.URL.Query().Get("redirect"),
		})
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	redirect := r.FormValue("redirect")

	token, _, err := h.auth.Login(email, password)
	if err != nil {
		var errMsg string
		if errors.Is(err, service.ErrInvalidCredentials) {
			errMsg = "Email ou mot de passe incorrect"
		} else if errors.Is(err, service.ErrAccountDisabled) {
			errMsg = "Votre compte a été désactivé. Contactez l'administration."
		} else {
			errMsg = "Une erreur est survenue. Veuillez réessayer."
		}
		h.render(w, r, "login.html", map[string]interface{}{
			"Error":    errMsg,
			"Email":    email,
			"Redirect": redirect,
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24h
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if redirect != "" && redirect[0] == '/' {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/tableau-de-bord", http.StatusSeeOther)
}

// RegisterHandler affiche le formulaire d'inscription
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/tableau-de-bord", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		h.render(w, r, "register.html", nil)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	if password != confirm {
		h.render(w, r, "register.html", map[string]interface{}{
			"Error":     "Les mots de passe ne correspondent pas",
			"FirstName": firstName,
			"LastName":  lastName,
			"Email":     email,
			"Phone":     phone,
		})
		return
	}

	if len(password) < 8 {
		h.render(w, r, "register.html", map[string]interface{}{
			"Error":     "Le mot de passe doit contenir au moins 8 caractères",
			"FirstName": firstName,
			"LastName":  lastName,
			"Email":     email,
			"Phone":     phone,
		})
		return
	}

	_, err := h.auth.Register(email, password, firstName, lastName, phone)
	if err != nil {
		errMsg := "Une erreur est survenue. Veuillez réessayer."
		if errors.Is(err, service.ErrEmailExists) {
			errMsg = "Cet email est déjà utilisé"
		}
		h.render(w, r, "register.html", map[string]interface{}{
			"Error":     errMsg,
			"FirstName": firstName,
			"LastName":  lastName,
			"Email":     email,
			"Phone":     phone,
		})
		return
	}

	redirectWithFlash(w, r, "/connexion", "success", "Compte créé avec succès ! Vous pouvez maintenant vous connecter.")
}

// LogoutHandler déconnecte l'utilisateur
func (h *Handlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
