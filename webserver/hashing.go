package webserver

import (
	"encoding/json"
	"errors"
	"log"
	"mazarin/firewall"
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

func LoadKeys(fileDir string) map[string]User {
	var usersMap = make(map[string]User)

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

	for _, users := range usersData.Users {
		_, ok := usersMap[users.Name]
		if ok {
			log.Printf("HASHING: Cant have two users named '%v' ", users.Name)
			return nil
		}
		usersMap[users.Name] = users
	}

	return usersMap
}

func HashKey(key string) (string, error) {
	if !firewall.ValidateInput(key, "password") {
		return "", errors.New("ERROR: Invalid password format. Only use letters,numbers and these symbols: _:/?#@!$&'()*+,;=- Between 12 and 64 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	return string(hash), err
}

func ValidateUserHash(password string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	return false, err
}
