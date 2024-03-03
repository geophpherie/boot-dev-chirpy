package database

import (
	"errors"
)

var ErrUserNotFound = errors.New("User not found")
var ErrUserAlreadyExists = errors.New("User already exists")

type User struct {
	Id             int    `json:"id"`
	Email          string `json:"email"`
	HashedPassword string `json:"password"`
}

func (db *DB) CreateUser(email string, hashedPassword string) (User, error) {
	// see if user exists
	if _, err := db.GetUserByEmail(email); !errors.Is(err, ErrUserNotFound) {
		// we could get ErrUserNotFound but that's okay, but return other errors
		// either a user exists or other error ocurred
		return User{}, ErrUserAlreadyExists
	}

	// get current Structure
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	// create new user
	userId := len(dbStructure.Users) + 1
	newUser := User{
		Id:             userId,
		Email:          email,
		HashedPassword: hashedPassword,
	}

	// add it to the db
	dbStructure.Users[userId] = newUser

	// dump the db (call write)
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (db *DB) UpdateUser(id int, email string, hashedPassword string) (User, error) {
	DbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	updatedUser := User{
		Id:             id,
		Email:          email,
		HashedPassword: hashedPassword,
	}

	DbStructure.Users[id] = updatedUser
	err = db.writeDB(DbStructure)
	if err != nil {
		return User{}, err
	}

	return updatedUser, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrUserNotFound
}

func (db *DB) GetUserById(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}

	return user, nil
}
