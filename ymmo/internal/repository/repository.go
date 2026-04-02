package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"ymmo/internal/models"
)

// convertPlaceholders converts ? placeholders to $1, $2, etc. for PostgreSQL
func convertPlaceholders(query string) string {
	var result strings.Builder
	placeholderCount := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			placeholderCount++
			fmt.Fprintf(&result, "$%d", placeholderCount)
		} else {
			result.WriteByte(query[i])
		}
	}
	return result.String()
}

// UserRepository gère les opérations CRUD sur les utilisateurs
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository crée un nouveau repository utilisateur
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID récupère un utilisateur par son ID
func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	u := &models.User{}
	var agencyID sql.NullInt64
	err := r.db.QueryRow(
		convertPlaceholders(`SELECT id, email, password_hash, first_name, last_name, phone, role, agency_id, avatar, is_active, created_at, updated_at
		 FROM users WHERE id = ? AND is_active = 1`), id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.Role, &agencyID, &u.Avatar, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if agencyID.Valid {
		u.AgencyID = &agencyID.Int64
	}
	return u, nil
}

// FindByEmail récupère un utilisateur par son email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	u := &models.User{}
	var agencyID sql.NullInt64
	err := r.db.QueryRow(
		convertPlaceholders(`SELECT id, email, password_hash, first_name, last_name, phone, role, agency_id, avatar, is_active, created_at, updated_at
		 FROM users WHERE email = ?`), email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.Role, &agencyID, &u.Avatar, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if agencyID.Valid {
		u.AgencyID = &agencyID.Int64
	}
	return u, nil
}

// Create crée un nouvel utilisateur
func (r *UserRepository) Create(u *models.User) error {
	res, err := r.db.Exec(
		convertPlaceholders(`INSERT INTO users (email, password_hash, first_name, last_name, phone, role, agency_id, is_active)
		 VALUES (?,?,?,?,?,?,?,1)`),
		u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone, u.Role, sqlNullInt64(u.AgencyID),
	)
	if err != nil {
		return err
	}
	u.ID, _ = res.LastInsertId()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// Update met à jour un utilisateur
func (r *UserRepository) Update(u *models.User) error {
	_, err := r.db.Exec(
		convertPlaceholders(`UPDATE users SET first_name=?, last_name=?, phone=?, avatar=?, updated_at=CURRENT_TIMESTAMP
		 WHERE id=?`),
		u.FirstName, u.LastName, u.Phone, u.Avatar, u.ID,
	)
	return err
}

// UpdatePassword met à jour le mot de passe
func (r *UserRepository) UpdatePassword(userID int64, hash string) error {
	_, err := r.db.Exec(
		convertPlaceholders("UPDATE users SET password_hash=?, updated_at=CURRENT_TIMESTAMP WHERE id=?"), hash, userID,
	)
	return err
}

// ListByAgency retourne les utilisateurs d'une agence
func (r *UserRepository) ListByAgency(agencyID int64) ([]*models.User, error) {
	rows, err := r.db.Query(
		convertPlaceholders(`SELECT id, email, first_name, last_name, phone, role, agency_id, avatar, is_active, created_at, updated_at
		 FROM users WHERE agency_id = ? AND role IN ('agent','director') ORDER BY last_name`),
		agencyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

// ListAll retourne tous les utilisateurs (admin)
func (r *UserRepository) ListAll(role string) ([]*models.User, error) {
	query := `SELECT id, email, first_name, last_name, phone, role, agency_id, avatar, is_active, created_at, updated_at
	          FROM users WHERE 1=1`
	args := []interface{}{}
	if role != "" {
		query += " AND role = ?"
		args = append(args, role)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.Query(convertPlaceholders(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

// Count retourne le nombre d'utilisateurs par rôle
func (r *UserRepository) Count(role string) (int, error) {
	query := "SELECT COUNT(*) FROM users WHERE is_active = 1"
	args := []interface{}{}
	if role != "" {
		query += " AND role = ?"
		args = append(args, role)
	}
	var count int
	return count, r.db.QueryRow(convertPlaceholders(query), args...).Scan(&count)
}

// SetRole met à jour le rôle d'un utilisateur
func (r *UserRepository) SetRole(userID int64, role models.Role, agencyID *int64) error {
	_, err := r.db.Exec(
		convertPlaceholders("UPDATE users SET role=?, agency_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?"),
		role, sqlNullInt64(agencyID), userID,
	)
	return err
}

// SetActive active ou désactive un compte
func (r *UserRepository) SetActive(userID int64, active bool) error {
	v := 0
	if active {
		v = 1
	}
	_, err := r.db.Exec(convertPlaceholders("UPDATE users SET is_active=? WHERE id=?"), v, userID)
	return err
}

func scanUsers(rows *sql.Rows) ([]*models.User, error) {
	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		var agencyID sql.NullInt64
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Phone,
			&u.Role, &agencyID, &u.Avatar, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if agencyID.Valid {
			u.AgencyID = &agencyID.Int64
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// -- PropertyRepository --

// PropertyRepository gère les opérations CRUD sur les biens
type PropertyRepository struct {
	db *sql.DB
}

// NewPropertyRepository crée un nouveau repository propriété
func NewPropertyRepository(db *sql.DB) *PropertyRepository {
	return &PropertyRepository{db: db}
}

// FindByID récupère un bien par son ID avec ses images
func (r *PropertyRepository) FindByID(id int64) (*models.Property, error) {
	p := &models.Property{}
	err := r.db.QueryRow(
		convertPlaceholders(`SELECT id, title, description, price, type, sub_type, status, surface, rooms, bedrooms, bathrooms,
		        floor, total_floors, garage, parking, garden, pool, elevator,
		        address, city, zip_code, department, latitude, longitude,
		        agency_id, agent_id, is_featured, created_at, updated_at
		 FROM properties WHERE id = ?`), id,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.Price, &p.Type, &p.SubType, &p.Status,
		&p.Surface, &p.Rooms, &p.Bedrooms, &p.Bathrooms,
		&p.Floor, &p.TotalFloors, &p.Garage, &p.Parking, &p.Garden, &p.Pool, &p.Elevator,
		&p.Address, &p.City, &p.ZipCode, &p.Department, &p.Latitude, &p.Longitude,
		&p.AgencyID, &p.AgentID, &p.IsFeatured, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	images, err := r.getImages(p.ID)
	if err != nil {
		return nil, err
	}
	p.Images = images
	return p, nil
}

// List retourne une liste paginée de biens avec filtres
func (r *PropertyRepository) List(f models.PropertyFilter) ([]*models.Property, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if f.Type != "" {
		where = append(where, "type = ?")
		args = append(args, f.Type)
	}
	if f.SubType != "" {
		where = append(where, "sub_type = ?")
		args = append(args, f.SubType)
	}
	if f.Status != "" {
		where = append(where, "status = ?")
		args = append(args, f.Status)
	}
	if f.City != "" {
		where = append(where, "LOWER(city) LIKE ?")
		args = append(args, "%"+strings.ToLower(f.City)+"%")
	}
	if f.Query != "" {
		where = append(where, "(LOWER(title) LIKE ? OR LOWER(city) LIKE ? OR LOWER(description) LIKE ?)")
		q := "%" + strings.ToLower(f.Query) + "%"
		args = append(args, q, q, q)
	}
	if f.MinPrice > 0 {
		where = append(where, "price >= ?")
		args = append(args, f.MinPrice)
	}
	if f.MaxPrice > 0 {
		where = append(where, "price <= ?")
		args = append(args, f.MaxPrice)
	}
	if f.MinSurface > 0 {
		where = append(where, "surface >= ?")
		args = append(args, f.MinSurface)
	}
	if f.MaxSurface > 0 {
		where = append(where, "surface <= ?")
		args = append(args, f.MaxSurface)
	}
	if f.MinRooms > 0 {
		where = append(where, "rooms >= ?")
		args = append(args, f.MinRooms)
	}
	if f.AgencyID > 0 {
		where = append(where, "agency_id = ?")
		args = append(args, f.AgencyID)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(
		convertPlaceholders("SELECT COUNT(*) FROM properties WHERE "+whereClause), args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit == 0 {
		f.Limit = 12
	}
	if f.Page < 1 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)

	rows, err := r.db.Query(
		convertPlaceholders(`SELECT id, title, description, price, type, sub_type, status, surface, rooms, bedrooms, bathrooms,
		        floor, total_floors, garage, parking, garden, pool, elevator,
		        address, city, zip_code, department, latitude, longitude,
		        agency_id, agent_id, is_featured, created_at, updated_at
		 FROM properties WHERE `+whereClause+
			` ORDER BY is_featured DESC, created_at DESC LIMIT ? OFFSET ?`),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var properties []*models.Property
	for rows.Next() {
		p := &models.Property{}
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Price, &p.Type, &p.SubType, &p.Status,
			&p.Surface, &p.Rooms, &p.Bedrooms, &p.Bathrooms,
			&p.Floor, &p.TotalFloors, &p.Garage, &p.Parking, &p.Garden, &p.Pool, &p.Elevator,
			&p.Address, &p.City, &p.ZipCode, &p.Department, &p.Latitude, &p.Longitude,
			&p.AgencyID, &p.AgentID, &p.IsFeatured, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		properties = append(properties, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Charger les images pour chaque bien
	for _, p := range properties {
		images, err := r.getImages(p.ID)
		if err != nil {
			return nil, 0, err
		}
		p.Images = images
	}

	return properties, total, nil
}

// ListFeatured retourne les biens mis en avant
func (r *PropertyRepository) ListFeatured(limit int) ([]*models.Property, error) {
	props, _, err := r.List(models.PropertyFilter{Page: 1, Limit: limit})
	return props, err
}

// Create crée un nouveau bien
func (r *PropertyRepository) Create(p *models.Property) error {
	res, err := r.db.Exec(
		convertPlaceholders(`INSERT INTO properties 
		 (title, description, price, type, sub_type, status, surface, rooms, bedrooms, bathrooms,
		  floor, total_floors, garage, parking, garden, pool, elevator,
		  address, city, zip_code, department, latitude, longitude,
		  agency_id, agent_id, is_featured)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`),
		p.Title, p.Description, p.Price, p.Type, p.SubType, p.Status,
		p.Surface, p.Rooms, p.Bedrooms, p.Bathrooms,
		p.Floor, p.TotalFloors, boolInt(p.Garage), boolInt(p.Parking),
		boolInt(p.Garden), boolInt(p.Pool), boolInt(p.Elevator),
		p.Address, p.City, p.ZipCode, p.Department, p.Latitude, p.Longitude,
		p.AgencyID, p.AgentID, boolInt(p.IsFeatured),
	)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

// Update met à jour un bien
func (r *PropertyRepository) Update(p *models.Property) error {
	_, err := r.db.Exec(
		convertPlaceholders(`UPDATE properties SET
		 title=?, description=?, price=?, type=?, sub_type=?, status=?,
		 surface=?, rooms=?, bedrooms=?, bathrooms=?, floor=?, total_floors=?,
		 garage=?, parking=?, garden=?, pool=?, elevator=?,
		 address=?, city=?, zip_code=?, department=?,
		 is_featured=?, updated_at=CURRENT_TIMESTAMP
		 WHERE id=?`),
		p.Title, p.Description, p.Price, p.Type, p.SubType, p.Status,
		p.Surface, p.Rooms, p.Bedrooms, p.Bathrooms, p.Floor, p.TotalFloors,
		boolInt(p.Garage), boolInt(p.Parking), boolInt(p.Garden), boolInt(p.Pool), boolInt(p.Elevator),
		p.Address, p.City, p.ZipCode, p.Department,
		boolInt(p.IsFeatured), p.ID,
	)
	return err
}

// Delete supprime un bien
func (r *PropertyRepository) Delete(id int64) error {
	_, err := r.db.Exec(convertPlaceholders("DELETE FROM properties WHERE id=?"), id)
	return err
}

// AddImage ajoute une image à un bien
func (r *PropertyRepository) AddImage(img *models.PropertyImage) error {
	res, err := r.db.Exec(
		convertPlaceholders("INSERT INTO property_images (property_id, url, is_primary, sort_order) VALUES (?,?,?,?)"),
		img.PropertyID, img.URL, boolInt(img.IsPrimary), img.SortOrder,
	)
	if err != nil {
		return err
	}
	img.ID, _ = res.LastInsertId()
	return nil
}

// DeleteImage supprime une image
func (r *PropertyRepository) DeleteImage(imageID int64) error {
	_, err := r.db.Exec(convertPlaceholders("DELETE FROM property_images WHERE id=?"), imageID)
	return err
}

// Count retourne le nombre de biens
func (r *PropertyRepository) Count(status string) (int, error) {
	query := "SELECT COUNT(*) FROM properties WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	var count int
	return count, r.db.QueryRow(convertPlaceholders(query), args...).Scan(&count)
}

// CountByAgency retourne le nombre de biens par agence
func (r *PropertyRepository) CountByAgency(agencyID int64) (int, error) {
	var count int
	return count, r.db.QueryRow(
		convertPlaceholders("SELECT COUNT(*) FROM properties WHERE agency_id = ?"), agencyID,
	).Scan(&count)
}

// ListByAgent retourne les biens d'un agent
func (r *PropertyRepository) ListByAgent(agentID int64) ([]*models.Property, error) {
	props, _, err := r.List(models.PropertyFilter{AgencyID: 0, Limit: 100, Page: 1})
	if err != nil {
		return nil, err
	}
	// Filtrer par agent
	var result []*models.Property
	for _, p := range props {
		if p.AgentID == agentID {
			result = append(result, p)
		}
	}
	return result, nil
}

// CityStats retourne les statistiques par ville
func (r *PropertyRepository) CityStats(limit int) ([]models.CityStats, error) {
	rows, err := r.db.Query(
		convertPlaceholders(`SELECT city, COUNT(*) as cnt, AVG(price) as avg_price
		 FROM properties GROUP BY city ORDER BY cnt DESC LIMIT ?`), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []models.CityStats
	for rows.Next() {
		s := models.CityStats{}
		if err := rows.Scan(&s.City, &s.Count, &s.AvgPrice); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// MonthlyStats retourne les statistiques mensuelles sur 12 mois
func (r *PropertyRepository) MonthlyStats() ([]models.MonthlyStats, error) {
	rows, err := r.db.Query(
		convertPlaceholders(`SELECT to_char(created_at, 'YYYY') as year, to_char(created_at, 'MM') as month, COUNT(*) as total
		 FROM properties
		 WHERE created_at >= CURRENT_DATE - INTERVAL '12 months'
		 GROUP BY to_char(created_at, 'YYYY'), to_char(created_at, 'MM') ORDER BY year, month`),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []models.MonthlyStats
	months := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}
	for rows.Next() {
		s := models.MonthlyStats{}
		var monthNum string
		if err := rows.Scan(&s.Year, &monthNum, &s.NewListings); err != nil {
			return nil, err
		}
		idx := 0
		fmt.Sscanf(monthNum, "%d", &idx)
		if idx >= 1 && idx <= 12 {
			s.Month = months[idx-1]
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *PropertyRepository) getImages(propertyID int64) ([]models.PropertyImage, error) {
	rows, err := r.db.Query(
		convertPlaceholders("SELECT id, property_id, url, is_primary, sort_order FROM property_images WHERE property_id = ? ORDER BY is_primary DESC, sort_order"),
		propertyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var images []models.PropertyImage
	for rows.Next() {
		img := models.PropertyImage{}
		var isPrimary int
		if err := rows.Scan(&img.ID, &img.PropertyID, &img.URL, &isPrimary, &img.SortOrder); err != nil {
			return nil, err
		}
		img.IsPrimary = isPrimary == 1
		images = append(images, img)
	}
	return images, rows.Err()
}

// -- FavoriteRepository --

// FavoriteRepository gère les favoris
type FavoriteRepository struct {
	db *sql.DB
}

// NewFavoriteRepository crée un nouveau repository favoris
func NewFavoriteRepository(db *sql.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Add ajoute un favori
func (r *FavoriteRepository) Add(userID, propertyID int64) error {
	_, err := r.db.Exec(
		convertPlaceholders("INSERT OR IGNORE INTO favorites (user_id, property_id) VALUES (?,?)"),
		userID, propertyID,
	)
	return err
}

// Remove supprime un favori
func (r *FavoriteRepository) Remove(userID, propertyID int64) error {
	_, err := r.db.Exec(
		convertPlaceholders("DELETE FROM favorites WHERE user_id=? AND property_id=?"),
		userID, propertyID,
	)
	return err
}

// IsFavorite vérifie si un bien est en favori
func (r *FavoriteRepository) IsFavorite(userID, propertyID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		convertPlaceholders("SELECT COUNT(*) FROM favorites WHERE user_id=? AND property_id=?"),
		userID, propertyID,
	).Scan(&count)
	return count > 0, err
}

// ListByUser retourne les favoris d'un utilisateur
func (r *FavoriteRepository) ListByUser(userID int64) ([]int64, error) {
	rows, err := r.db.Query(
		convertPlaceholders("SELECT property_id FROM favorites WHERE user_id=?"), userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// -- ContactRepository --

// ContactRepository gère les demandes de contact
type ContactRepository struct {
	db *sql.DB
}

// NewContactRepository crée un nouveau repository contact
func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create crée une demande de contact
func (r *ContactRepository) Create(c *models.ContactRequest) error {
	res, err := r.db.Exec(
		convertPlaceholders(`INSERT INTO contact_requests (user_id, property_id, agent_id, full_name, email, phone, message, status)
		 VALUES (?,?,?,?,?,?,?,'pending')`),
		sqlNullInt64(c.UserID), c.PropertyID, c.AgentID, c.FullName, c.Email, c.Phone, c.Message,
	)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

// ListByAgent retourne les demandes d'un agent
func (r *ContactRepository) ListByAgent(agentID int64) ([]*models.ContactRequest, error) {
	rows, err := r.db.Query(
		convertPlaceholders(`SELECT id, user_id, property_id, agent_id, full_name, email, phone, message, status, created_at
		 FROM contact_requests WHERE agent_id=? ORDER BY created_at DESC`), agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContacts(rows)
}

// ListAll retourne toutes les demandes (admin)
func (r *ContactRepository) ListAll() ([]*models.ContactRequest, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, property_id, agent_id, full_name, email, phone, message, status, created_at
		 FROM contact_requests ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContacts(rows)
}

// UpdateStatus met à jour le statut d'une demande
func (r *ContactRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(convertPlaceholders("UPDATE contact_requests SET status=? WHERE id=?"), status, id)
	return err
}

// Count retourne le nombre de demandes
func (r *ContactRepository) Count() (int, error) {
	var count int
	return count, r.db.QueryRow("SELECT COUNT(*) FROM contact_requests").Scan(&count)
}

func scanContacts(rows *sql.Rows) ([]*models.ContactRequest, error) {
	var contacts []*models.ContactRequest
	for rows.Next() {
		c := &models.ContactRequest{}
		var userID sql.NullInt64
		if err := rows.Scan(&c.ID, &userID, &c.PropertyID, &c.AgentID,
			&c.FullName, &c.Email, &c.Phone, &c.Message, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		if userID.Valid {
			c.UserID = &userID.Int64
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

// -- AgencyRepository --

// AgencyRepository gère les opérations sur les agences
type AgencyRepository struct {
	db *sql.DB
}

// NewAgencyRepository crée un nouveau repository agence
func NewAgencyRepository(db *sql.DB) *AgencyRepository {
	return &AgencyRepository{db: db}
}

// FindByID récupère une agence par ID
func (r *AgencyRepository) FindByID(id int64) (*models.Agency, error) {
	a := &models.Agency{}
	err := r.db.QueryRow(
		convertPlaceholders(`SELECT id, name, city, address, phone, email, website, description, logo, created_at
		 FROM agencies WHERE id = ?`), id,
	).Scan(&a.ID, &a.Name, &a.City, &a.Address, &a.Phone, &a.Email, &a.Website, &a.Description, &a.Logo, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ListAll retourne toutes les agences
func (r *AgencyRepository) ListAll() ([]*models.Agency, error) {
	rows, err := r.db.Query(
		`SELECT id, name, city, address, phone, email, website, description, logo, created_at
		 FROM agencies ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var agencies []*models.Agency
	for rows.Next() {
		a := &models.Agency{}
		if err := rows.Scan(&a.ID, &a.Name, &a.City, &a.Address, &a.Phone, &a.Email, &a.Website, &a.Description, &a.Logo, &a.CreatedAt); err != nil {
			return nil, err
		}
		agencies = append(agencies, a)
	}
	return agencies, rows.Err()
}

// Create crée une nouvelle agence
func (r *AgencyRepository) Create(a *models.Agency) error {
	res, err := r.db.Exec(
		convertPlaceholders(`INSERT INTO agencies (name, city, address, phone, email, website, description, logo)
		 VALUES (?,?,?,?,?,?,?,?)`),
		a.Name, a.City, a.Address, a.Phone, a.Email, a.Website, a.Description, a.Logo,
	)
	if err != nil {
		return err
	}
	a.ID, _ = res.LastInsertId()
	return nil
}

// Update met à jour une agence
func (r *AgencyRepository) Update(a *models.Agency) error {
	_, err := r.db.Exec(
		convertPlaceholders(`UPDATE agencies SET name=?, city=?, address=?, phone=?, email=?, website=?, description=?, logo=?
		 WHERE id=?`),
		a.Name, a.City, a.Address, a.Phone, a.Email, a.Website, a.Description, a.Logo, a.ID,
	)
	return err
}

// Delete supprime une agence
func (r *AgencyRepository) Delete(id int64) error {
	_, err := r.db.Exec(convertPlaceholders("DELETE FROM agencies WHERE id=?"), id)
	return err
}

// Count retourne le nombre d'agences
func (r *AgencyRepository) Count() (int, error) {
	var count int
	return count, r.db.QueryRow("SELECT COUNT(*) FROM agencies").Scan(&count)
}

// -- Helpers --

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func sqlNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}
