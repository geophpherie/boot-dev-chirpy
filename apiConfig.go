package main

import (
	"github.com/jbeyer16/boot-dev-chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
	jwtSecret      string
}
