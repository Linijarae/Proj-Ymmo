package config

import (
	"os"
)

type Config struct {
	Port       string
	DBPath     string
	JWTSecret  string
	Env        string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./ymmo.db"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "ymmo-super-secret-key-change-in-production"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	return &Config{
		Port:      port,
		DBPath:    dbPath,
		JWTSecret: jwtSecret,
		Env:       env,
	}
}
