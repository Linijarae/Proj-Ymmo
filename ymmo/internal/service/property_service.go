package service

import (
	"ymmo/internal/models"
	"ymmo/internal/repository"
)

// PropertyService gère la logique métier des biens immobiliers
type PropertyService struct {
	properties *repository.PropertyRepository
	agencies   *repository.AgencyRepository
	users      *repository.UserRepository
	favorites  *repository.FavoriteRepository
	contacts   *repository.ContactRepository
}

// NewPropertyService crée un nouveau service de gestion des biens
func NewPropertyService(
	props *repository.PropertyRepository,
	agencies *repository.AgencyRepository,
	users *repository.UserRepository,
	favorites *repository.FavoriteRepository,
	contacts *repository.ContactRepository,
) *PropertyService {
	return &PropertyService{
		properties: props,
		agencies:   agencies,
		users:      users,
		favorites:  favorites,
		contacts:   contacts,
	}
}

// GetProperty retourne un bien avec toutes ses relations chargées
func (s *PropertyService) GetProperty(id int64) (*models.Property, error) {
	p, err := s.properties.FindByID(id)
	if err != nil || p == nil {
		return p, err
	}

	// Charger l'agence
	agency, err := s.agencies.FindByID(p.AgencyID)
	if err == nil && agency != nil {
		p.Agency = agency
	}

	// Charger l'agent
	agent, err := s.users.FindByID(p.AgentID)
	if err == nil && agent != nil {
		p.Agent = agent
	}

	return p, nil
}

// ListProperties retourne une liste paginée de biens avec filtres
func (s *PropertyService) ListProperties(f models.PropertyFilter) ([]*models.Property, int, error) {
	props, total, err := s.properties.List(f)
	if err != nil {
		return nil, 0, err
	}
	// Enrichir avec les relations agences
	for _, p := range props {
		agency, err := s.agencies.FindByID(p.AgencyID)
		if err == nil && agency != nil {
			p.Agency = agency
		}
	}
	return props, total, nil
}

// ListFeatured retourne les biens vedettes pour la home
func (s *PropertyService) ListFeatured(limit int) ([]*models.Property, error) {
	f := models.PropertyFilter{Page: 1, Limit: limit}
	props, _, err := s.properties.List(f)
	if err != nil {
		return nil, err
	}
	return props, nil
}

// CreateProperty crée un nouveau bien (par un agent/admin)
func (s *PropertyService) CreateProperty(p *models.Property) error {
	return s.properties.Create(p)
}

// UpdateProperty met à jour un bien
func (s *PropertyService) UpdateProperty(p *models.Property, userID int64, role models.Role) error {
	existing, err := s.properties.FindByID(p.ID)
	if err != nil || existing == nil {
		return err
	}
	// Vérification des droits
	if role != models.RoleSuperAdmin && existing.AgentID != userID {
		return models.ErrForbidden
	}
	return s.properties.Update(p)
}

// DeleteProperty supprime un bien
func (s *PropertyService) DeleteProperty(id, userID int64, role models.Role) error {
	existing, err := s.properties.FindByID(id)
	if err != nil || existing == nil {
		return err
	}
	if role != models.RoleSuperAdmin && existing.AgentID != userID {
		return models.ErrForbidden
	}
	return s.properties.Delete(id)
}

// ToggleFavorite ajoute/supprime un bien des favoris
func (s *PropertyService) ToggleFavorite(userID, propertyID int64) (bool, error) {
	isFav, err := s.favorites.IsFavorite(userID, propertyID)
	if err != nil {
		return false, err
	}
	if isFav {
		return false, s.favorites.Remove(userID, propertyID)
	}
	return true, s.favorites.Add(userID, propertyID)
}

// GetUserFavoriteIDs retourne les IDs des biens favoris d'un utilisateur
func (s *PropertyService) GetUserFavoriteIDs(userID int64) ([]int64, error) {
	return s.favorites.ListByUser(userID)
}

// GetUserFavorites retourne les biens favoris complets d'un utilisateur
func (s *PropertyService) GetUserFavorites(userID int64) ([]*models.Property, error) {
	ids, err := s.favorites.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	var result []*models.Property
	for _, id := range ids {
		p, err := s.GetProperty(id)
		if err == nil && p != nil {
			result = append(result, p)
		}
	}
	return result, nil
}

// SendContactRequest envoie une demande de contact pour un bien
func (s *PropertyService) SendContactRequest(req *models.ContactRequest) error {
	return s.contacts.Create(req)
}

// GetAgentContacts retourne les demandes de contact d'un agent
func (s *PropertyService) GetAgentContacts(agentID int64) ([]*models.ContactRequest, error) {
	contacts, err := s.contacts.ListByAgent(agentID)
	if err != nil {
		return nil, err
	}
	// Enrichir avec les biens
	for _, c := range contacts {
		p, err := s.properties.FindByID(c.PropertyID)
		if err == nil && p != nil {
			c.Property = p
		}
	}
	return contacts, nil
}

// UpdateContactStatus met à jour le statut d'une demande
func (s *PropertyService) UpdateContactStatus(id int64, status string) error {
	return s.contacts.UpdateStatus(id, status)
}

// GetAgentProperties retourne les biens gérés par un agent
func (s *PropertyService) GetAgentProperties(agentID int64) ([]*models.Property, error) {
	return s.properties.ListByAgent(agentID)
}

// AnalyticsSummary retourne un résumé statistique
func (s *PropertyService) GetAnalyticsSummary() (*models.AnalyticsSummary, error) {
	summary := &models.AnalyticsSummary{}

	counts := map[string]*int{
		"for_sale": &summary.TotalForSale,
		"for_rent": &summary.TotalForRent,
		"sold":     &summary.TotalSold,
		"rented":   &summary.TotalRented,
	}
	for status, ptr := range counts {
		n, err := s.properties.Count(status)
		if err != nil {
			return nil, err
		}
		*ptr = n
		summary.TotalProperties += n
	}

	total, err := s.agencies.Count()
	if err != nil {
		return nil, err
	}
	summary.TotalAgencies = total

	agents, err := s.users.Count("agent")
	if err != nil {
		return nil, err
	}
	summary.TotalAgents = agents

	clients, err := s.users.Count("client")
	if err != nil {
		return nil, err
	}
	summary.TotalClients = clients

	contacts, err := s.contacts.Count()
	if err != nil {
		return nil, err
	}
	summary.TotalContacts = contacts

	return summary, nil
}

// GetCityStats retourne les statistiques par ville
func (s *PropertyService) GetCityStats(limit int) ([]models.CityStats, error) {
	return s.properties.CityStats(limit)
}

// GetMonthlyStats retourne les statistiques mensuelles
func (s *PropertyService) GetMonthlyStats() ([]models.MonthlyStats, error) {
	return s.properties.MonthlyStats()
}
