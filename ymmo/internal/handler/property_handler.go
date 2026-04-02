package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"ymmo/internal/middleware"
	"ymmo/internal/models"
)

// PropertiesHandler affiche la liste des biens avec filtres
func (h *Handlers) PropertiesHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	minPrice, _ := strconv.ParseFloat(q.Get("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(q.Get("max_price"), 64)
	minSurface, _ := strconv.ParseFloat(q.Get("min_surface"), 64)
	minRooms, _ := strconv.Atoi(q.Get("min_rooms"))

	f := models.PropertyFilter{
		Query:      q.Get("q"),
		Type:       q.Get("type"),
		SubType:    q.Get("sub_type"),
		Status:     q.Get("status"),
		City:       q.Get("city"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		MinSurface: minSurface,
		MinRooms:   minRooms,
		Page:       page,
		Limit:      12,
	}

	properties, total, err := h.property.ListProperties(f)
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(f.Limit)))
	agencies, _ := h.agencies.ListAll()

	h.render(w, r, "properties.html", map[string]interface{}{
		"Properties":  properties,
		"Filter":      f,
		"Total":       total,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"Agencies":    agencies,
	})
}

// PropertyDetailHandler affiche le détail d'un bien
func (h *Handlers) PropertyDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	property, err := h.property.GetProperty(id)
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	if property == nil {
		http.NotFound(w, r)
		return
	}

	var isFavorite bool
	claims := middleware.GetClaims(r)
	if claims != nil {
		ids, _ := h.property.GetUserFavoriteIDs(claims.UserID)
		for _, fid := range ids {
			if fid == id {
				isFavorite = true
				break
			}
		}
	}

	// Biens similaires (même ville, même type)
	similar, _, _ := h.property.ListProperties(models.PropertyFilter{
		City:  property.City,
		Type:  string(property.Type),
		Limit: 3,
		Page:  1,
	})
	// Retirer le bien actuel des similaires
	var filteredSimilar []*models.Property
	for _, s := range similar {
		if s.ID != property.ID {
			filteredSimilar = append(filteredSimilar, s)
		}
	}

	h.render(w, r, "property.html", map[string]interface{}{
		"Property":   property,
		"IsFavorite": isFavorite,
		"Similar":    filteredSimilar,
	})
}

// PropertyCreateHandler affiche le formulaire de création d'un bien
func (h *Handlers) PropertyCreateHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	agencies, _ := h.agencies.ListAll()

	if r.Method == http.MethodGet {
		h.render(w, r, "property-form.html", map[string]interface{}{
			"Property": &models.Property{},
			"Agencies": agencies,
			"IsNew":    true,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulaire invalide", http.StatusBadRequest)
		return
	}

	p := parsePropertyForm(r)
	p.AgentID = claims.UserID
	if claims.AgencyID != nil {
		p.AgencyID = *claims.AgencyID
	} else {
		agencyID, _ := strconv.ParseInt(r.FormValue("agency_id"), 10, 64)
		p.AgencyID = agencyID
	}

	if err := h.property.CreateProperty(p); err != nil {
		h.render(w, r, "property-form.html", map[string]interface{}{
			"Error":    "Erreur lors de la création du bien",
			"Property": p,
			"Agencies": agencies,
			"IsNew":    true,
		})
		return
	}

	redirectWithFlash(w, r, fmt.Sprintf("/biens/%d", p.ID), "success", "Bien créé avec succès !")
}

// PropertyEditHandler affiche le formulaire d'édition d'un bien
func (h *Handlers) PropertyEditHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	property, err := h.property.GetProperty(id)
	if err != nil || property == nil {
		http.NotFound(w, r)
		return
	}

	agencies, _ := h.agencies.ListAll()

	if r.Method == http.MethodGet {
		h.render(w, r, "property-form.html", map[string]interface{}{
			"Property": property,
			"Agencies": agencies,
			"IsNew":    false,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulaire invalide", http.StatusBadRequest)
		return
	}

	updated := parsePropertyForm(r)
	updated.ID = id
	updated.AgentID = property.AgentID
	updated.AgencyID = property.AgencyID

	if err := h.property.UpdateProperty(updated, claims.UserID, claims.Role); err != nil {
		if err == models.ErrForbidden {
			http.Error(w, "Accès refusé", http.StatusForbidden)
			return
		}
		h.render(w, r, "property-form.html", map[string]interface{}{
			"Error":    "Erreur lors de la mise à jour",
			"Property": updated,
			"Agencies": agencies,
			"IsNew":    false,
		})
		return
	}

	redirectWithFlash(w, r, fmt.Sprintf("/biens/%d", id), "success", "Bien mis à jour avec succès !")
}

// PropertyDeleteHandler supprime un bien
func (h *Handlers) PropertyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.property.DeleteProperty(id, claims.UserID, claims.Role); err != nil {
		if err == models.ErrForbidden {
			http.Error(w, "Accès refusé", http.StatusForbidden)
			return
		}
		setFlash(w, "error", "Erreur lors de la suppression")
		http.Redirect(w, r, "/tableau-de-bord", http.StatusSeeOther)
		return
	}

	redirectWithFlash(w, r, "/tableau-de-bord", "success", "Bien supprimé avec succès")
}

// FavoriteToggleHandler ajoute/supprime un favori
func (h *Handlers) FavoriteToggleHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	added, err := h.property.ToggleFavorite(claims.UserID, id)
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	if added {
		w.Write([]byte("added"))
	} else {
		w.Write([]byte("removed"))
	}
}

// ContactHandler envoie une demande de contact
func (h *Handlers) ContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulaire invalide", http.StatusBadRequest)
		return
	}

	propertyID, err := strconv.ParseInt(r.FormValue("property_id"), 10, 64)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	agentID, err := strconv.ParseInt(r.FormValue("agent_id"), 10, 64)
	if err != nil {
		http.Error(w, "Agent invalide", http.StatusBadRequest)
		return
	}

	req := &models.ContactRequest{
		PropertyID: propertyID,
		AgentID:    agentID,
		FullName:   r.FormValue("full_name"),
		Email:      r.FormValue("email"),
		Phone:      r.FormValue("phone"),
		Message:    r.FormValue("message"),
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		req.UserID = &claims.UserID
	}

	if err := h.property.SendContactRequest(req); err != nil {
		setFlash(w, "error", "Erreur lors de l'envoi de votre demande")
		http.Redirect(w, r, fmt.Sprintf("/biens/%d", propertyID), http.StatusSeeOther)
		return
	}

	redirectWithFlash(w, r, fmt.Sprintf("/biens/%d", propertyID), "success",
		"Votre demande a bien été envoyée ! L'agent vous contactera prochainement.")
}

// AgenciesHandler affiche la liste des agences
func (h *Handlers) AgenciesHandler(w http.ResponseWriter, r *http.Request) {
	agencies, err := h.agencies.ListAll()
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	h.render(w, r, "agencies.html", map[string]interface{}{
		"Agencies": agencies,
	})
}

// AgencyDetailHandler affiche le détail d'une agence
func (h *Handlers) AgencyDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	agency, err := h.agencies.FindByID(id)
	if err != nil || agency == nil {
		http.NotFound(w, r)
		return
	}

	properties, _, _ := h.property.ListProperties(models.PropertyFilter{
		AgencyID: id,
		Page:     1,
		Limit:    8,
	})

	agents, _ := h.usersRepo.ListAll("")
	var agencyAgents []*models.User
	for _, u := range agents {
		if u.AgencyID != nil && *u.AgencyID == id && (u.Role == models.RoleAgent || u.Role == models.RoleDirector) {
			agencyAgents = append(agencyAgents, u)
		}
	}

	h.render(w, r, "agency.html", map[string]interface{}{
		"Agency":     agency,
		"Properties": properties,
		"Agents":     agencyAgents,
	})
}

// parsePropertyForm parse le formulaire de création/édition d'un bien
func parsePropertyForm(r *http.Request) *models.Property {
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	surface, _ := strconv.ParseFloat(r.FormValue("surface"), 64)
	rooms, _ := strconv.Atoi(r.FormValue("rooms"))
	bedrooms, _ := strconv.Atoi(r.FormValue("bedrooms"))
	bathrooms, _ := strconv.Atoi(r.FormValue("bathrooms"))
	floor, _ := strconv.Atoi(r.FormValue("floor"))
	totalFloors, _ := strconv.Atoi(r.FormValue("total_floors"))

	return &models.Property{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Price:       price,
		Type:        models.PropertyType(r.FormValue("type")),
		SubType:     models.PropertySubType(r.FormValue("sub_type")),
		Status:      models.PropertyStatus(r.FormValue("status")),
		Surface:     surface,
		Rooms:       rooms,
		Bedrooms:    bedrooms,
		Bathrooms:   bathrooms,
		Floor:       floor,
		TotalFloors: totalFloors,
		Garage:      r.FormValue("garage") == "on",
		Parking:     r.FormValue("parking") == "on",
		Garden:      r.FormValue("garden") == "on",
		Pool:        r.FormValue("pool") == "on",
		Elevator:    r.FormValue("elevator") == "on",
		Address:     r.FormValue("address"),
		City:        r.FormValue("city"),
		ZipCode:     r.FormValue("zip_code"),
		Department:  r.FormValue("department"),
		IsFeatured:  r.FormValue("is_featured") == "on",
	}
}
