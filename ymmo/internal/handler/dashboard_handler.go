package handler

import (
	"net/http"
	"strconv"
	"ymmo/internal/middleware"
	"ymmo/internal/models"
)

// DashboardHandler affiche le tableau de bord utilisateur
func (h *Handlers) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{}

	switch claims.Role {
	case models.RoleSuperAdmin:
		// Données admin
		summary, _ := h.property.GetAnalyticsSummary()
		allUsers, _ := h.usersRepo.ListAll("")
		allAgencies, _ := h.agencies.ListAll()
		allContacts, _ := h.contacts.ListAll()
		recentProps, _, _ := h.property.ListProperties(models.PropertyFilter{Page: 1, Limit: 10})
		data["Summary"] = summary
		data["Users"] = allUsers
		data["Agencies"] = allAgencies
		data["Contacts"] = allContacts
		data["RecentProperties"] = recentProps

	case models.RoleDirector:
		// Données directeur d'agence
		if claims.AgencyID != nil {
			agency, _ := h.agencies.FindByID(*claims.AgencyID)
			props, _, _ := h.property.ListProperties(models.PropertyFilter{
				AgencyID: *claims.AgencyID,
				Page:     1,
				Limit:    20,
			})
			agents, _ := h.usersRepo.ListAll("")
			var agencyAgents []*models.User
			for _, u := range agents {
				if u.AgencyID != nil && *u.AgencyID == *claims.AgencyID {
					agencyAgents = append(agencyAgents, u)
				}
			}
			data["Agency"] = agency
			data["Properties"] = props
			data["Agents"] = agencyAgents
		}

	case models.RoleAgent:
		// Données agent
		myProps, _ := h.property.GetAgentProperties(claims.UserID)
		myContacts, _ := h.property.GetAgentContacts(claims.UserID)
		data["Properties"] = myProps
		data["Contacts"] = myContacts

	case models.RoleClient:
		// Données client
		favorites, _ := h.property.GetUserFavorites(claims.UserID)
		data["Favorites"] = favorites
	}

	h.render(w, r, "dashboard.html", data)
}

// ProfileHandler affiche et met à jour le profil
func (h *Handlers) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	u, err := h.usersRepo.FindByID(claims.UserID)
	if err != nil || u == nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		h.render(w, r, "profile.html", map[string]interface{}{
			"Profile": u,
		})
		return
	}

	u.FirstName = r.FormValue("first_name")
	u.LastName = r.FormValue("last_name")
	u.Phone = r.FormValue("phone")

	if err := h.usersRepo.Update(u); err != nil {
		h.render(w, r, "profile.html", map[string]interface{}{
			"Error":   "Erreur lors de la mise à jour",
			"Profile": u,
		})
		return
	}

	redirectWithFlash(w, r, "/profil", "success", "Profil mis à jour avec succès !")
}

// AdminUsersHandler gère les utilisateurs (admin)
func (h *Handlers) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil || claims.Role != models.RoleSuperAdmin {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		userIDStr := r.FormValue("user_id")
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)

		switch action {
		case "set_role":
			roleStr := r.FormValue("role")
			agencyIDStr := r.FormValue("agency_id")
			var agencyID *int64
			if agencyIDStr != "" {
				if id, err := strconv.ParseInt(agencyIDStr, 10, 64); err == nil && id > 0 {
					agencyID = &id
				}
			}
			h.usersRepo.SetRole(userID, models.Role(roleStr), agencyID)
		case "toggle_active":
			u, _ := h.usersRepo.FindByID(userID)
			if u != nil {
				h.usersRepo.SetActive(userID, !u.IsActive)
			}
		}

		redirectWithFlash(w, r, "/admin/utilisateurs", "success", "Modification effectuée")
		return
	}

	users, _ := h.usersRepo.ListAll("")
	agencies, _ := h.agencies.ListAll()
	h.render(w, r, "admin-users.html", map[string]interface{}{
		"Users":    users,
		"Agencies": agencies,
	})
}

// AdminAgenciesHandler gère les agences (admin)
func (h *Handlers) AdminAgenciesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil || (claims.Role != models.RoleSuperAdmin && claims.Role != models.RoleDirector) {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		agencyIDStr := r.FormValue("agency_id")
		agencyID, _ := strconv.ParseInt(agencyIDStr, 10, 64)

		switch action {
		case "create":
			a := &models.Agency{
				Name:        r.FormValue("name"),
				City:        r.FormValue("city"),
				Address:     r.FormValue("address"),
				Phone:       r.FormValue("phone"),
				Email:       r.FormValue("email"),
				Description: r.FormValue("description"),
			}
			h.agencies.Create(a)
		case "delete":
			if claims.Role == models.RoleSuperAdmin {
				h.agencies.Delete(agencyID)
			}
		}

		redirectWithFlash(w, r, "/admin/agences", "success", "Modification effectuée")
		return
	}

	agencies, _ := h.agencies.ListAll()
	h.render(w, r, "admin-agencies.html", map[string]interface{}{
		"Agencies": agencies,
	})
}

// AdminContactsHandler gère les demandes de contact (admin/agent)
func (h *Handlers) AdminContactsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		contactIDStr := r.FormValue("contact_id")
		status := r.FormValue("status")
		contactID, _ := strconv.ParseInt(contactIDStr, 10, 64)
		h.contacts.UpdateStatus(contactID, status)
		redirectWithFlash(w, r, "/admin/contacts", "success", "Statut mis à jour")
		return
	}

	var contacts []*models.ContactRequest
	if claims.Role == models.RoleSuperAdmin || claims.Role == models.RoleDirector {
		contacts, _ = h.contacts.ListAll()
	} else {
		contacts, _ = h.property.GetAgentContacts(claims.UserID)
	}

	h.render(w, r, "admin-contacts.html", map[string]interface{}{
		"Contacts": contacts,
	})
}

// AnalyticsHandler affiche les statistiques
func (h *Handlers) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	summary, _ := h.property.GetAnalyticsSummary()
	cityStats, _ := h.property.GetCityStats(10)
	monthlyStats, _ := h.property.GetMonthlyStats()

	h.render(w, r, "analytics.html", map[string]interface{}{
		"Summary":      summary,
		"CityStats":    cityStats,
		"MonthlyStats": monthlyStats,
	})
}
