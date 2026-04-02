package models

import "time"

// PropertyType définit la catégorie du bien
type PropertyType string

const (
	PropertyTypeResidential PropertyType = "residential"
	PropertyTypeCommercial  PropertyType = "commercial"
)

// PropertySubType définit le sous-type du bien
type PropertySubType string

const (
	// Résidentiel
	SubTypeApartment PropertySubType = "apartment"
	SubTypeHouse     PropertySubType = "house"
	SubTypeVilla     PropertySubType = "villa"
	SubTypeStudio    PropertySubType = "studio"
	SubTypeLoft      PropertySubType = "loft"
	SubTypeDuplex    PropertySubType = "duplex"
	SubTypeTerraced  PropertySubType = "terraced"
	// Commercial
	SubTypeOffice    PropertySubType = "office"
	SubTypeRetail    PropertySubType = "retail"
	SubTypeWarehouse PropertySubType = "warehouse"
	SubTypeBuilding  PropertySubType = "building"
	SubTypeLand      PropertySubType = "land"
)

// PropertyStatus définit le statut du bien
type PropertyStatus string

const (
	StatusForSale  PropertyStatus = "for_sale"
	StatusForRent  PropertyStatus = "for_rent"
	StatusSold     PropertyStatus = "sold"
	StatusRented   PropertyStatus = "rented"
)

// Property représente un bien immobilier
type Property struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Type        PropertyType   `json:"type"`
	SubType     PropertySubType `json:"sub_type"`
	Status      PropertyStatus `json:"status"`
	Surface     float64        `json:"surface"`
	Rooms       int            `json:"rooms"`
	Bedrooms    int            `json:"bedrooms"`
	Bathrooms   int            `json:"bathrooms"`
	Floor       int            `json:"floor"`
	TotalFloors int            `json:"total_floors"`
	Garage      bool           `json:"garage"`
	Parking     bool           `json:"parking"`
	Garden      bool           `json:"garden"`
	Pool        bool           `json:"pool"`
	Elevator    bool           `json:"elevator"`
	Address     string         `json:"address"`
	City        string         `json:"city"`
	ZipCode     string         `json:"zip_code"`
	Department  string         `json:"department"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	AgencyID    int64          `json:"agency_id"`
	AgentID     int64          `json:"agent_id"`
	IsFeatured  bool           `json:"is_featured"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Relations (chargées à la demande)
	Images  []PropertyImage `json:"images,omitempty"`
	Agency  *Agency         `json:"agency,omitempty"`
	Agent   *User           `json:"agent,omitempty"`
}

// PrimaryImage retourne l'image principale du bien
func (p *Property) PrimaryImage() string {
	for _, img := range p.Images {
		if img.IsPrimary {
			return img.URL
		}
	}
	if len(p.Images) > 0 {
		return p.Images[0].URL
	}
	return "/static/img/property-default.jpg"
}

// IsAvailable vérifie si le bien est disponible
func (p *Property) IsAvailable() bool {
	return p.Status == StatusForSale || p.Status == StatusForRent
}

// FormattedPrice retourne le prix formaté
func (p *Property) StatusLabel() string {
	switch p.Status {
	case StatusForSale:
		return "À vendre"
	case StatusForRent:
		return "À louer"
	case StatusSold:
		return "Vendu"
	case StatusRented:
		return "Loué"
	default:
		return string(p.Status)
	}
}

// SubTypeLabel retourne le libellé du sous-type
func (p *Property) SubTypeLabel() string {
	labels := map[PropertySubType]string{
		SubTypeApartment: "Appartement",
		SubTypeHouse:     "Maison",
		SubTypeVilla:     "Villa",
		SubTypeStudio:    "Studio",
		SubTypeLoft:      "Loft",
		SubTypeDuplex:    "Duplex",
		SubTypeTerraced:  "Maison de ville",
		SubTypeOffice:    "Bureau",
		SubTypeRetail:    "Commerce",
		SubTypeWarehouse: "Entrepôt",
		SubTypeBuilding:  "Immeuble",
		SubTypeLand:      "Terrain",
	}
	if label, ok := labels[p.SubType]; ok {
		return label
	}
	return string(p.SubType)
}

// PropertyImage représente une image d'un bien
type PropertyImage struct {
	ID         int64  `json:"id"`
	PropertyID int64  `json:"property_id"`
	URL        string `json:"url"`
	IsPrimary  bool   `json:"is_primary"`
	SortOrder  int    `json:"sort_order"`
}

// PropertyFilter définit les filtres de recherche
type PropertyFilter struct {
	Query      string
	Type       string
	SubType    string
	Status     string
	City       string
	MinPrice   float64
	MaxPrice   float64
	MinSurface float64
	MaxSurface float64
	MinRooms   int
	AgencyID   int64
	Page       int
	Limit      int
}

// ContactRequest représente une demande de contact
type ContactRequest struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"user_id,omitempty"`
	PropertyID int64     `json:"property_id"`
	AgentID    int64     `json:"agent_id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Message    string    `json:"message"`
	Status     string    `json:"status"` // pending, processed, closed
	CreatedAt  time.Time `json:"created_at"`

	// Relations
	Property *Property `json:"property,omitempty"`
	Agent    *User     `json:"agent,omitempty"`
}

// Favorite représente un bien mis en favori
type Favorite struct {
	UserID     int64     `json:"user_id"`
	PropertyID int64     `json:"property_id"`
	CreatedAt  time.Time `json:"created_at"`

	Property *Property `json:"property,omitempty"`
}
