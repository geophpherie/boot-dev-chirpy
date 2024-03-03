package database

import (
	"errors"
	"fmt"
	"time"
)

var ErrTokenRevoked = errors.New("token has been revoked")
var ErrTokenNotFound = errors.New("token not present")

func (db *DB) AddToken(token string) error {
	// get current Structure
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	// add the token
	dbStructure.Tokens[token] = time.Time{}

	// dump the db (call write)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CheckToken(token string) error {
	// get current Structure
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	dat, ok := dbStructure.Tokens[token]
	if !ok {
		return ErrTokenNotFound
	}

	if dat != (time.Time{}) {
		fmt.Printf("Token revoked at %v", dat)
		return ErrTokenRevoked
	}

	return nil
}

func (db *DB) RevokeToken(token string) error {
	// get current Structure
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	_, ok := dbStructure.Tokens[token]
	if !ok {
		return ErrTokenNotFound
	}

	dbStructure.Tokens[token] = time.Now()

	// dump the db (call write)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}
