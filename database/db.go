package database

import (
	"database/sql"
	"fmt"
	"mazarin/config"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type User struct {
	ID                int
	Username          string
	PasswordHash      string
	CreatedAt         time.Time
	PermissionGroupID int
	Active            bool
}

var currentDB *sql.DB = nil

func InitDb(conf *config.WebserverConfig) error {
	DbDir := conf.DbDir
	err := os.MkdirAll(DbDir, os.ModePerm) // Create db dir if it doesn't exist
	if err != nil {
		return fmt.Errorf("failed to create db directory: %v", err)
	}

	dbFilePath := filepath.Join(DbDir, "mazarinDB")

	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return err
	}
	currentDB = db

	if err := setupDb(); err != nil {
		return fmt.Errorf("failed to setup db: %v", err)
	}

	return nil
}

func setupDb() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	schema := `
    -- Users table (migrating from keys.json)
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        permission_group_id INTEGER,
        active BOOLEAN DEFAULT 1
    );
    CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`

	_, err := db.Exec(schema)
	return err
}

func GetDB() *sql.DB {
	return currentDB
}

func CreateUser(username, passwordHash string, groupID int) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(`
        INSERT INTO users (username, password_hash, permission_group_id) 
        VALUES (?, ?, ?)
    `, username, passwordHash, groupID)

	return err
}

func GetUserByUsername(username string) (*User, error) {
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	user := &User{}
	err := db.QueryRow(`
        SELECT id, username, password_hash, created_at, updated_at, permission_group_id, active 
        FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.PermissionGroupID, &user.Active)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func UpdateUser(userID int, passwordHash string, groupID int, active bool) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(`
        UPDATE users 
        SET password_hash = ?, permission_group_id = ?, active = ?
        WHERE id = ?
    `, passwordHash, groupID, active, userID)

	return err
}

func DeleteUser(userID int) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	return err
}

func ListUsers() ([]User, error) {
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(`
        SELECT id, username, password_hash, created_at, updated_at, 
               COALESCE(permission_group_id, 1), active 
        FROM users ORDER BY username
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash,
			&user.CreatedAt, &user.PermissionGroupID, &user.Active,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
