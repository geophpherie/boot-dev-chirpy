package database

import "errors"

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

func (db *DB) CreateChirp(body string, author_id int) (Chirp, error) {
	// get current Structure
	DbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	// create new chirp
	chirpId := len(DbStructure.Chirps) + 1
	newChirp := Chirp{
		Id:       chirpId,
		Body:     body,
		AuthorId: author_id,
	}

	// add it to the db
	DbStructure.Chirps[chirpId] = newChirp

	// dump the db (call write)
	err = db.writeDB(DbStructure)
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

	chirps := make([]Chirp, 0, len(DbStructure.Chirps))
	for _, v := range DbStructure.Chirps {
		chirps = append(chirps, v)
	}

	return chirps, nil

}

func (db *DB) DeleteChirp(chirpId int) error {
	DbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	_, ok := DbStructure.Chirps[chirpId]
	if !ok {
		return errors.New("chirp not found")
	}

	delete(DbStructure.Chirps, chirpId)

	return nil
}
