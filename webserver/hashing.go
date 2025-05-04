package webserver

import (
	"encoding/json"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name            string `json:"name"`
	Hash            string `json:"hash"`
	AllowedSessions int    `json:"allowed_sessions"`
}

type UsersData struct {
	Users []User `json:"users"`
}

func LoadKeys(fileDir string) *UsersData {
	data, err := os.ReadFile(fileDir + "/keys.json")
	if err != nil {
		log.Println("HASHING: LoadKeys error ", err)
		return nil
	}

	var usersData UsersData
	err = json.Unmarshal(data, &usersData)
	if err != nil {
		log.Println("HASHING: Unmarshal error ", err)
		return nil
	}

	return &usersData
}

func HashKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	return string(hash), err
}

func validateUserHash(password string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	return false, err
}
