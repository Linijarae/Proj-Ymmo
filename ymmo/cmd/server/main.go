package main

import (
	"log"
	"net/http"
	"ymmo/internal/config"
	"ymmo/internal/database"
	"ymmo/internal/handler"
	"ymmo/internal/middleware"
	"ymmo/internal/repository"
	"ymmo/internal/service"
)

func main() {
	cfg := config.Load()

	// Initialisation de la base de données
	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Impossible d'initialiser la base de données: %v", err)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db.DB)
	propertyRepo := repository.NewPropertyRepository(db.DB)
	agencyRepo := repository.NewAgencyRepository(db.DB)
	favoriteRepo := repository.NewFavoriteRepository(db.DB)
	contactRepo := repository.NewContactRepository(db.DB)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	propertySvc := service.NewPropertyService(propertyRepo, agencyRepo, userRepo, favoriteRepo, contactRepo)

	// Handlers
	h := handler.NewHandlers(authSvc, propertySvc, userRepo, agencyRepo, contactRepo)

	// Routeur
	mux := http.NewServeMux()

	// Fichiers statiques
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Pages publiques
	mux.HandleFunc("GET /{$}", h.HomeHandler)
	mux.HandleFunc("GET /biens", h.PropertiesHandler)
	mux.HandleFunc("GET /biens/{id}", h.PropertyDetailHandler)
	mux.HandleFunc("GET /agences", h.AgenciesHandler)
	mux.HandleFunc("GET /agences/{id}", h.AgencyDetailHandler)
	mux.HandleFunc("GET /statistiques", h.AnalyticsHandler)

	// Authentification
	mux.HandleFunc("GET /connexion", h.LoginHandler)
	mux.HandleFunc("POST /connexion", h.LoginHandler)
	mux.HandleFunc("GET /inscription", h.RegisterHandler)
	mux.HandleFunc("POST /inscription", h.RegisterHandler)
	mux.HandleFunc("GET /deconnexion", h.LogoutHandler)

	// Zone protégée
	mux.Handle("GET /tableau-de-bord", middleware.RequireAuth(http.HandlerFunc(h.DashboardHandler)))
	mux.Handle("GET /profil", middleware.RequireAuth(http.HandlerFunc(h.ProfileHandler)))
	mux.Handle("POST /profil", middleware.RequireAuth(http.HandlerFunc(h.ProfileHandler)))

	// Gestion des biens (agents/admin)
	mux.Handle("GET /biens/nouveau",
		middleware.RequireRole(
			"super_admin", "director", "agent",
		)(http.HandlerFunc(h.PropertyCreateHandler)))
	mux.Handle("POST /biens/nouveau",
		middleware.RequireRole(
			"super_admin", "director", "agent",
		)(http.HandlerFunc(h.PropertyCreateHandler)))
	mux.Handle("GET /biens/{id}/modifier",
		middleware.RequireAuth(http.HandlerFunc(h.PropertyEditHandler)))
	mux.Handle("POST /biens/{id}/modifier",
		middleware.RequireAuth(http.HandlerFunc(h.PropertyEditHandler)))
	mux.Handle("POST /biens/{id}/supprimer",
		middleware.RequireAuth(http.HandlerFunc(h.PropertyDeleteHandler)))

	// Contact et favoris
	mux.HandleFunc("POST /contact", h.ContactHandler)
	mux.Handle("POST /favoris/{id}",
		middleware.RequireAuth(http.HandlerFunc(h.FavoriteToggleHandler)))

	// Administration
	mux.Handle("GET /admin/utilisateurs",
		middleware.RequireRole("super_admin")(http.HandlerFunc(h.AdminUsersHandler)))
	mux.Handle("POST /admin/utilisateurs",
		middleware.RequireRole("super_admin")(http.HandlerFunc(h.AdminUsersHandler)))
	mux.Handle("GET /admin/agences",
		middleware.RequireRole("super_admin", "director")(http.HandlerFunc(h.AdminAgenciesHandler)))
	mux.Handle("POST /admin/agences",
		middleware.RequireRole("super_admin", "director")(http.HandlerFunc(h.AdminAgenciesHandler)))
	mux.Handle("GET /admin/contacts",
		middleware.RequireAuth(http.HandlerFunc(h.AdminContactsHandler)))
	mux.Handle("POST /admin/contacts",
		middleware.RequireAuth(http.HandlerFunc(h.AdminContactsHandler)))

	// Chaîne de middlewares globaux
	chain := middleware.SecurityHeaders(
		middleware.CSRF(
			middleware.Auth(authSvc)(mux),
		),
	)

	log.Printf("🏠 Ymmo démarré sur http://localhost:%s", cfg.Port)
	log.Printf("📧 Admin: admin@ymmo.fr | Mot de passe: Password123!")
	log.Printf("👤 Client: client@example.com | Mot de passe: Password123!")
	if err := http.ListenAndServe(":"+cfg.Port, chain); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
