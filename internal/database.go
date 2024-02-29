package internal

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

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	// get current Structure
	DbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	// create new chirp
	chirpId := len(DbStructure.Chirps) + 1
	newChirp := Chirp{Id: chirpId, Body: body}

	// add it to the db
	DbStructure.Chirps[chirpId] = newChirp

	// dump the db (call write)
	err = db.writeDB(DbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

func (db *DB) CreateUser(email string) (User, error) {
	// get current Structure
	DbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	// create new chirp
	userId := len(DbStructure.Users) + 1
	newUser := User{Id: userId, Email: email}

	// add it to the db
	DbStructure.Users[userId] = newUser

	// dump the db (call write)
	err = db.writeDB(DbStructure)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}
func (db *DB) GetChirp(chirpId int) (Chirp, error) {
	DbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	Chirp, ok := DbStructure.Chirps[chirpId]
	if !ok {
		return Chirp, errors.New("chirp not found")
	}
	return Chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	DbStructure, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}

	if len(DbStructure.Chirps) == 0 {
		return []Chirp{}, nil
	}

	chirps := make([]Chirp, 0, len(DbStructure.Chirps))
	for _, v := range DbStructure.Chirps {
		chirps = append(chirps, v)
	}

	return chirps, nil

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

	DbStructure := DBStructure{}
	err = json.Unmarshal(data, &DbStructure)
	if err != nil {
		return DBStructure{}, err
	}
	return DbStructure, nil
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
	DB := &DB{path: path, mux: &sync.RWMutex{}}

	err := DB.ensureDB()

	return DB, err
}
