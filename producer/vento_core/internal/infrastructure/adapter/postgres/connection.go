package postgres

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id         VARCHAR(36)  PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   TEXT         NOT NULL,
    full_name  VARCHAR(255) NOT NULL,
    business_name VARCHAR(255) NOT NULL DEFAULT '',
    google_id  VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='business_name') THEN
        ALTER TABLE users ADD COLUMN business_name VARCHAR(255) NOT NULL DEFAULT '';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='google_id') THEN
        ALTER TABLE users ADD COLUMN google_id VARCHAR(255) UNIQUE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`

// NewConnection opens a PostgreSQL connection pool and runs migrations.
func NewConnection(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Run auto-migration
	if _, err := db.Exec(createUsersTable); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ PostgreSQL connected and migrations applied")
	return db, nil
}
