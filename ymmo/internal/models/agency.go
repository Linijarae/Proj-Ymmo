package models

import "time"

// Agency représente une agence immobilière du réseau Ymmo
type Agency struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	City        string    `json:"city"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Website     string    `json:"website"`
	Description string    `json:"description"`
	Logo        string    `json:"logo"`
	CreatedAt   time.Time `json:"created_at"`

	// Relations
	Director       *User  `json:"director,omitempty"`
	AgentCount     int    `json:"agent_count,omitempty"`
	PropertyCount  int    `json:"property_count,omitempty"`
}

// AnalyticsSummary contient les statistiques globales
type AnalyticsSummary struct {
	TotalProperties   int     `json:"total_properties"`
	TotalForSale      int     `json:"total_for_sale"`
	TotalForRent      int     `json:"total_for_rent"`
	TotalSold         int     `json:"total_sold"`
	TotalRented       int     `json:"total_rented"`
	TotalAgencies     int     `json:"total_agencies"`
	TotalAgents       int     `json:"total_agents"`
	TotalClients      int     `json:"total_clients"`
	TotalContacts     int     `json:"total_contacts"`
	AvgPriceSale      float64 `json:"avg_price_sale"`
	AvgPriceRent      float64 `json:"avg_price_rent"`
	MostActiveCity    string  `json:"most_active_city"`
}

// CityStats représente les statistiques par ville
type CityStats struct {
	City          string  `json:"city"`
	Count         int     `json:"count"`
	AvgPrice      float64 `json:"avg_price"`
}

// MonthlyStats représente les statistiques mensuelles
type MonthlyStats struct {
	Month     string `json:"month"`
	Year      int    `json:"year"`
	NewListings int  `json:"new_listings"`
	Sold        int  `json:"sold"`
	Rented      int  `json:"rented"`
}

// PriceRange représente une tranche de prix
type PriceRange struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}
