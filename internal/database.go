package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

var chirpId = 0

func (db *DB) CreateChirp(body string) (Chirp, error) {
	// get current chirps
	Chirps, err := db.GetChirps()
	if err != nil {
		fmt.Println("Unable to read Chirps")
		return Chirp{}, err
	}

	// create new chirp
	chirpId++
	newChirp := Chirp{Id: chirpId, Body: body}
	if len(Chirps) == 0 {
		Chirps = []Chirp{newChirp}
	} else {
		Chirps = append(Chirps, newChirp)
	}

	// write all chirps into the DB Structure
	newDbStructure := DBStructure{Chirps: map[int]Chirp{}}
	for _, chirp := range Chirps {
		id := chirp.Id
		newDbStructure.Chirps[id] = chirp
	}
	// dump the db (call write)
	err = db.writeDB(newDbStructure)

	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil

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

	var Chirps []Chirp

	for _, v := range DbStructure.Chirps {
		Chirps = append(Chirps, v)
	}

	return Chirps, nil

}
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(db.path, []byte("{}"), 0666)
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
