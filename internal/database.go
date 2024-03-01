package internal

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
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
	Id             int    `json:"id"`
	Email          string `json:"email"`
	HashedPassword string `json:"password"`
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

func (db *DB) CreateUser(email string, password string) (User, error) {
	// see if user exists
	_, err := db.GetUser(email)

	// intentional because no error means they were findable
	if err == nil {
		return User{}, errors.New("User already exists")
	}

	// get current Structure
	DbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	// create new user
	userId := len(DbStructure.Users) + 1

	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	newUser := User{Id: userId, Email: email, HashedPassword: string(hashedPassword)}

	// add it to the db
	DbStructure.Users[userId] = newUser

	// dump the db (call write)
	err = db.writeDB(DbStructure)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	DbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	updatedUser := User{Id: id, Email: email, HashedPassword: string(hashedPassword)}
	DbStructure.Users[id] = updatedUser

	err = db.writeDB(DbStructure)
	if err != nil {
		return User{}, err
	}

	return updatedUser, nil
}

func (db *DB) GetUser(email string) (User, error) {
	DbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range DbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, errors.New("User not found")
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
