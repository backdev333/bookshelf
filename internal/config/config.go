package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
}

func Load() (*Config, error) {
	var err error

	portEnv := os.Getenv("PORT")
	port := 8080
	if portEnv != "" {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			return nil, fmt.Errorf("can't parse by atoi port: %w", err)
		}
	}

	dbUrlEnv := os.Getenv("DATABASE_URL")
	dbUrl := "postgresql://username:password@localhost:5432/database"
	if dbUrlEnv != "" {
		dbUrl = dbUrlEnv
	}

	jwtEnv := os.Getenv("JWT_SECRET")
	jwtSecret := "0a6876f139eea1103c5d74dd72f09c92fe8d00e4806bb6f8006d596a139b506c"
	if jwtEnv != "" {
		jwtSecret = jwtEnv
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbUrl,
		JWTSecret:   jwtSecret,
	}, nil
}
