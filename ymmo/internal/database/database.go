package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// DB est le wrapper autour de *sql.DB
type DB struct {
	*sql.DB
}

// New ouvre la base de données SQLite et crée les tables si nécessaire
func New(dsn string) (*DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base de données: %w", err)
	}

	// Activer les clés étrangères et WAL pour de meilleures performances
	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;"); err != nil {
		return nil, fmt.Errorf("configuration PRAGMA: %w", err)
	}

	wrapper := &DB{db}
	if err := wrapper.migrate(); err != nil {
		return nil, fmt.Errorf("migration: %w", err)
	}

	log.Println("✅ Base de données initialisée")
	return wrapper, nil
}

// migrate crée le schéma de la base de données
func (db *DB) migrate() error {
	schema := `
	-- Table des agences
	CREATE TABLE IF NOT EXISTS agencies (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL,
		city        TEXT NOT NULL,
		address     TEXT NOT NULL DEFAULT '',
		phone       TEXT NOT NULL DEFAULT '',
		email       TEXT NOT NULL DEFAULT '',
		website     TEXT DEFAULT '',
		description TEXT DEFAULT '',
		logo        TEXT DEFAULT '',
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Table des utilisateurs
	CREATE TABLE IF NOT EXISTS users (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		email         TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		first_name    TEXT NOT NULL DEFAULT '',
		last_name     TEXT NOT NULL DEFAULT '',
		phone         TEXT DEFAULT '',
		role          TEXT NOT NULL DEFAULT 'client' CHECK(role IN ('super_admin','director','agent','client')),
		agency_id     INTEGER REFERENCES agencies(id) ON DELETE SET NULL,
		avatar        TEXT DEFAULT '',
		is_active     INTEGER NOT NULL DEFAULT 1,
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Table des biens immobiliers
	CREATE TABLE IF NOT EXISTS properties (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		title        TEXT NOT NULL,
		description  TEXT DEFAULT '',
		price        REAL NOT NULL DEFAULT 0,
		type         TEXT NOT NULL DEFAULT 'residential' CHECK(type IN ('residential','commercial')),
		sub_type     TEXT NOT NULL DEFAULT 'apartment',
		status       TEXT NOT NULL DEFAULT 'for_sale' CHECK(status IN ('for_sale','for_rent','sold','rented')),
		surface      REAL DEFAULT 0,
		rooms        INTEGER DEFAULT 0,
		bedrooms     INTEGER DEFAULT 0,
		bathrooms    INTEGER DEFAULT 0,
		floor        INTEGER DEFAULT 0,
		total_floors INTEGER DEFAULT 0,
		garage       INTEGER DEFAULT 0,
		parking      INTEGER DEFAULT 0,
		garden       INTEGER DEFAULT 0,
		pool         INTEGER DEFAULT 0,
		elevator     INTEGER DEFAULT 0,
		address      TEXT DEFAULT '',
		city         TEXT NOT NULL DEFAULT '',
		zip_code     TEXT DEFAULT '',
		department   TEXT DEFAULT '',
		latitude     REAL DEFAULT 0,
		longitude    REAL DEFAULT 0,
		agency_id    INTEGER NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
		agent_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		is_featured  INTEGER DEFAULT 0,
		created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Table des images des biens
	CREATE TABLE IF NOT EXISTS property_images (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
		url         TEXT NOT NULL,
		is_primary  INTEGER DEFAULT 0,
		sort_order  INTEGER DEFAULT 0
	);

	-- Table des favoris
	CREATE TABLE IF NOT EXISTS favorites (
		user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, property_id)
	);

	-- Table des demandes de contact
	CREATE TABLE IF NOT EXISTS contact_requests (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER REFERENCES users(id) ON DELETE SET NULL,
		property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
		agent_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		full_name   TEXT NOT NULL DEFAULT '',
		email       TEXT NOT NULL DEFAULT '',
		phone       TEXT DEFAULT '',
		message     TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','processed','closed')),
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Index pour les recherches fréquentes
	CREATE INDEX IF NOT EXISTS idx_properties_city    ON properties(city);
	CREATE INDEX IF NOT EXISTS idx_properties_status  ON properties(status);
	CREATE INDEX IF NOT EXISTS idx_properties_type    ON properties(type);
	CREATE INDEX IF NOT EXISTS idx_properties_agency  ON properties(agency_id);
	CREATE INDEX IF NOT EXISTS idx_properties_price   ON properties(price);
	CREATE INDEX IF NOT EXISTS idx_users_email        ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_agency       ON users(agency_id);
	CREATE INDEX IF NOT EXISTS idx_contacts_agent     ON contact_requests(agent_id);
	CREATE INDEX IF NOT EXISTS idx_contacts_property  ON contact_requests(property_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return db.seed()
}

// seed insère des données de démonstration si la base est vide
func (db *DB) seed() error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // déjà peuplé
	}

	log.Println("🌱 Insertion des données de démonstration...")

	// Hash bcrypt pour "Password123!" (généré à l'avance pour le seed)
	// En production, utiliser le service auth pour créer les utilisateurs
	pwHash := "$2a$12$1AwpjxNUy1bt8lWF31QRe.wt22cHYKjgBXQzd9nmRKkpCxp1d3IPO"

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Agences
	agencies := []struct{ name, city, address, phone, email string }{
		{"Ymmo Aix-en-Provence (Siège)", "Aix-en-Provence", "12 Cours Mirabeau, 13100 Aix-en-Provence", "04 42 00 00 00", "siege@ymmo.fr"},
		{"Ymmo Paris 8ème", "Paris", "45 Avenue des Champs-Élysées, 75008 Paris", "01 40 00 00 01", "paris8@ymmo.fr"},
		{"Ymmo Lyon Part-Dieu", "Lyon", "18 Rue de la Part-Dieu, 69003 Lyon", "04 72 00 00 02", "lyon@ymmo.fr"},
		{"Ymmo Marseille Vieux-Port", "Marseille", "3 Quai du Port, 13002 Marseille", "04 91 00 00 03", "marseille@ymmo.fr"},
		{"Ymmo Nice Côte d'Azur", "Nice", "10 Promenade des Anglais, 06000 Nice", "04 93 00 00 04", "nice@ymmo.fr"},
		{"Ymmo Bordeaux Chartrons", "Bordeaux", "25 Quai des Chartrons, 33000 Bordeaux", "05 56 00 00 05", "bordeaux@ymmo.fr"},
		{"Ymmo Toulouse Capitole", "Toulouse", "8 Place du Capitole, 31000 Toulouse", "05 34 00 00 06", "toulouse@ymmo.fr"},
		{"Ymmo Montpellier Comédie", "Montpellier", "5 Place de la Comédie, 34000 Montpellier", "04 67 00 00 07", "montpellier@ymmo.fr"},
		{"Ymmo Nantes Île de Nantes", "Nantes", "2 Boulevard de la Prairie au Duc, 44000 Nantes", "02 40 00 00 08", "nantes@ymmo.fr"},
		{"Ymmo Strasbourg Neustadt", "Strasbourg", "15 Avenue de la Liberté, 67000 Strasbourg", "03 88 00 00 09", "strasbourg@ymmo.fr"},
		{"Ymmo Lille Grand Place", "Lille", "7 Grand Place, 59000 Lille", "03 20 00 00 10", "lille@ymmo.fr"},
		{"Ymmo Rennes Vilaine", "Rennes", "30 Quai de la Vilaine, 35000 Rennes", "02 99 00 00 11", "rennes@ymmo.fr"},
		{"Ymmo Grenoble Alpes", "Grenoble", "12 Rue Félix Viallet, 38000 Grenoble", "04 76 00 00 12", "grenoble@ymmo.fr"},
	}

	agencyIDs := make([]int64, 0)
	for _, a := range agencies {
		res, err := tx.Exec(
			"INSERT INTO agencies (name, city, address, phone, email, description) VALUES (?,?,?,?,?,?)",
			a.name, a.city, a.address, a.phone, a.email,
			"Agence Ymmo spécialisée dans la vente et la location de biens résidentiels et commerciaux.",
		)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		agencyIDs = append(agencyIDs, id)
	}

	// Super Admin
	res, err := tx.Exec(
		`INSERT INTO users (email, password_hash, first_name, last_name, phone, role, is_active)
		 VALUES (?,?,?,?,?,?,1)`,
		"admin@ymmo.fr", pwHash, "Marie", "Dupont", "04 42 00 00 00", "super_admin",
	)
	if err != nil {
		return err
	}
	adminID, _ := res.LastInsertId()

	// Directeur siège
	res, err = tx.Exec(
		`INSERT INTO users (email, password_hash, first_name, last_name, phone, role, agency_id, is_active)
		 VALUES (?,?,?,?,?,?,?,1)`,
		"directeur@ymmo.fr", pwHash, "Pierre", "Martin", "04 42 00 00 01", "director", agencyIDs[0],
	)
	if err != nil {
		return err
	}

	// Agents commerciaux
	agents := []struct {
		firstName, lastName, email string
		agencyID                   int64
	}{
		{"Sophie", "Leblanc", "s.leblanc@ymmo.fr", agencyIDs[0]},
		{"Thomas", "Bernard", "t.bernard@ymmo.fr", agencyIDs[1]},
		{"Julie", "Moreau", "j.moreau@ymmo.fr", agencyIDs[2]},
		{"Nicolas", "Laurent", "n.laurent@ymmo.fr", agencyIDs[3]},
		{"Camille", "Simon", "c.simon@ymmo.fr", agencyIDs[4]},
	}

	agentIDs := make([]int64, 0)
	for _, ag := range agents {
		r, err := tx.Exec(
			`INSERT INTO users (email, password_hash, first_name, last_name, role, agency_id, is_active)
			 VALUES (?,?,?,?,?,?,1)`,
			ag.email, pwHash, ag.firstName, ag.lastName, "agent", ag.agencyID,
		)
		if err != nil {
			return err
		}
		id, _ := r.LastInsertId()
		agentIDs = append(agentIDs, id)
	}

	// Client de démonstration
	_, err = tx.Exec(
		`INSERT INTO users (email, password_hash, first_name, last_name, phone, role, is_active)
		 VALUES (?,?,?,?,?,?,1)`,
		"client@example.com", pwHash, "Jean", "Durand", "06 12 34 56 78", "client",
	)
	if err != nil {
		return err
	}

	_ = adminID

	// Biens immobiliers de démonstration
	properties := []struct {
		title, description, ptype, subtype, status, city, zip, address string
		price, surface                                                  float64
		rooms, bedrooms, bathrooms                                      int
		agency_id, agent_id                                             int64
		featured                                                        bool
		garden, pool, garage, parking                                   bool
	}{
		{
			"Magnifique Villa avec Piscine", "Superbe villa contemporaine avec piscine à débordement, vue panoramique et finitions haut de gamme. Jardin paysager de 1500m².",
			"residential", "villa", "for_sale", "Aix-en-Provence", "13100", "Chemin des Vignes",
			1250000, 280, 6, 4, 3, agencyIDs[0], agentIDs[0], true, true, true, true, true,
		},
		{
			"Appartement T3 Centre-Ville Nice", "Bel appartement lumineux en plein cœur de Nice, proche de la Promenade des Anglais. Vue mer depuis le salon.",
			"residential", "apartment", "for_sale", "Nice", "06000", "Rue Masséna",
			485000, 75, 3, 2, 1, agencyIDs[4], agentIDs[4], true, false, false, false, true,
		},
		{
			"Maison de Ville Bordeaux", "Maison de caractère avec cour intérieure dans le quartier des Chartrons. Rénovée avec goût, alliant charme de l'ancien et modernité.",
			"residential", "house", "for_sale", "Bordeaux", "33000", "Rue Notre-Dame",
			620000, 165, 5, 3, 2, agencyIDs[5], agentIDs[1], true, true, false, true, false,
		},
		{
			"Studio Meublé Paris 8ème", "Studio moderne et entièrement meublé à deux pas des Champs-Élysées. Idéal pour un premier investissement ou une résidence secondaire.",
			"residential", "studio", "for_rent", "Paris", "75008", "Rue du Faubourg Saint-Honoré",
			1800, 28, 1, 0, 1, agencyIDs[1], agentIDs[1], false, false, false, false, false,
		},
		{
			"Loft Industriel Lyon Confluence", "Loft atypique dans un ancien entrepôt réhabilité. Hauts plafonds, grandes baies vitrées, esprit industriel chic.",
			"residential", "loft", "for_sale", "Lyon", "69002", "Quai Rambaud",
			390000, 120, 2, 1, 1, agencyIDs[2], agentIDs[2], false, false, false, false, true,
		},
		{
			"Bureau Moderne Open Space Marseille", "Plateau de bureaux en open space, idéalement situé dans le quartier d'affaires de Marseille. Fibre optique et espaces communs.",
			"commercial", "office", "for_rent", "Marseille", "13008", "Avenue du Prado",
			4500, 180, 0, 0, 2, agencyIDs[3], agentIDs[3], false, false, false, false, true,
		},
		{
			"Duplex Terrasse Toulouse", "Magnifique duplex avec grande terrasse ensoleillée offrant une vue dégagée sur la ville rose. Belle luminosité toute la journée.",
			"residential", "duplex", "for_sale", "Toulouse", "31000", "Allées Jean Jaurès",
			340000, 95, 4, 2, 2, agencyIDs[6], agentIDs[4], false, true, false, false, true,
		},
		{
			"Maison Familiale Nantes", "Grande maison familiale dans un quartier calme et résidentiel de Nantes. Proche des écoles et commerces. Jardin arboré.",
			"residential", "house", "for_sale", "Nantes", "44000", "Rue des Jardins",
			420000, 145, 6, 4, 2, agencyIDs[8], agentIDs[0], false, true, false, true, true,
		},
		{
			"Appartement Haussmannien Paris 8ème", "Somptueux appartement haussmannien entièrement rénové. Parquet en chêne massif, moulures d'époque, cuisine équipée haut de gamme.",
			"residential", "apartment", "for_sale", "Paris", "75008", "Avenue Hoche",
			1850000, 185, 5, 3, 2, agencyIDs[1], agentIDs[1], true, false, false, false, false,
		},
		{
			"Local Commercial Centre Strasbourg", "Local commercial idéalement situé en hyper-centre de Strasbourg. Fort passage piéton, vitrine sur rue.",
			"commercial", "retail", "for_rent", "Strasbourg", "67000", "Rue des Grandes Arcades",
			3200, 85, 0, 0, 1, agencyIDs[9], agentIDs[2], false, false, false, false, false,
		},
		{
			"Villa Contemporaine Montpellier", "Villa contemporaine de plain-pied avec piscine et jardin paysager. Architecture moderne, matériaux nobles.",
			"residential", "villa", "for_sale", "Montpellier", "34000", "Domaine du Golf",
			890000, 210, 5, 4, 3, agencyIDs[7], agentIDs[3], true, true, true, true, true,
		},
		{
			"Appartement T2 Grenoble", "Appartement T2 lumineux avec balcon, vue sur le Vercors. Résidence sécurisée avec gardien, parking en sous-sol.",
			"residential", "apartment", "for_rent", "Grenoble", "38000", "Avenue du Grésivaudan",
			750, 48, 2, 1, 1, agencyIDs[12], agentIDs[4], false, false, false, false, true,
		},
	}

	for _, p := range properties {
		r, err := tx.Exec(
			`INSERT INTO properties 
			 (title, description, price, type, sub_type, status, surface, rooms, bedrooms, bathrooms, 
			  address, city, zip_code, agency_id, agent_id, is_featured, garden, pool, garage, parking, elevator)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			p.title, p.description, p.price, p.ptype, p.subtype, p.status,
			p.surface, p.rooms, p.bedrooms, p.bathrooms,
			p.address, p.city, p.zip, p.agency_id, p.agent_id, p.featured,
			p.garden, p.pool, p.garage, p.parking, false,
		)
		if err != nil {
			return err
		}
		propID, _ := r.LastInsertId()

		// Image principale fictive (SVG placeholder)
		_, err = tx.Exec(
			"INSERT INTO property_images (property_id, url, is_primary, sort_order) VALUES (?,?,1,0)",
			propID, fmt.Sprintf("/static/img/property-%d.svg", (propID%5)+1),
		)
		if err != nil {
			return err
		}
	}

	// Demandes de contact de démonstration
	contactsData := []struct {
		propID          int64
		agentID         int64
		name, email, msg string
	}{
		{1, agentIDs[0], "Jean Durand", "client@example.com", "Je suis intéressé par cette villa. Pouvez-vous me contacter pour une visite ?"},
		{2, agentIDs[4], "Marie Leclerc", "m.leclerc@email.com", "Bonjour, est-il possible de visiter cet appartement ce week-end ?"},
		{9, agentIDs[1], "Paul Rousseau", "p.rousseau@email.com", "Je recherche un appartement haussmannien. Celui-ci m'intéresse beaucoup."},
	}
	for _, c := range contactsData {
		_, err = tx.Exec(
			`INSERT INTO contact_requests (property_id, agent_id, full_name, email, message, status)
			 VALUES (?,?,?,?,?,?)`,
			c.propID, c.agentID, c.name, c.email, c.msg, "pending",
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("✅ Données de démonstration insérées")
	return nil
}
