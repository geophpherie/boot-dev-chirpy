package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		newDb := DBStructure{Chirps: map[int]Chirp{}, Users: map[int]User{}}
		err = db.writeDB(newDb)
	}

	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	dbStructure := DBStructure{}
	err = json.Unmarshal(data, &dbStructure)
	if err != nil {
		return DBStructure{}, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0666)
	if err != nil {
		return err
	}

	return nil
}

func NewDB(path string) (*DB, error) {
	db := &DB{path: path, mux: &sync.RWMutex{}}

	err := db.ensureDB()

	return db, err
}
