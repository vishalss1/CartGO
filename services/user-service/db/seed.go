package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/vishalss1/CartGO/pkg/util"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdminUser ensures a default admin user exists in the database.
// This is called during service startup.
func SeedAdminUser(db *sql.DB) error {
	// Check if admin already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "admin@cartgo.com").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Println("Seeding admin user...")

	// Generate UUID using unified generator
	id := util.GenerateUUID()

	// Hash password
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert admin user
	query := `
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (email) DO NOTHING`

	now := time.Now()
	_, err = db.Exec(query,
		id,
		"admin",
		"admin@cartgo.com",
		string(hash),
		"ADMIN",
		now,
		now,
	)

	if err != nil {
		return err
	}

	log.Println("Admin user seeded successfully")
	return nil
}
