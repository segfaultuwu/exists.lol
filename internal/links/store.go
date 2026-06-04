package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Domain struct {
	Subdomain  string
	RecordType string
	Value      string
	Status     string
}

type User struct {
	DiscordID      string
	DiscordName    string
	GitHubUsername string
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}

	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS users (
			discord_id TEXT PRIMARY KEY,
			discord_name TEXT NOT NULL,
			github_username TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			discord_id TEXT NOT NULL,
			subdomain TEXT NOT NULL UNIQUE,
			record_type TEXT NOT NULL,
			value TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (discord_id) REFERENCES users(discord_id)
				ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_domains_discord_id
			ON domains(discord_id);

		CREATE INDEX IF NOT EXISTS idx_domains_subdomain
			ON domains(subdomain);
	`)
	return err
}

func (s *Store) IsAlreadyLinked(githubName string, discordId string) bool {
	if githubName == "" {
		return false
	}

	row := s.db.QueryRow(`
		SELECT discord_id
		FROM users
		WHERE LOWER(github_username) = LOWER(?)
		LIMIT 1;
	`, githubName)

	var existingDiscordID string

	err := row.Scan(&existingDiscordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}

		// If the database returns an error, fail closed.
		// This prevents accidentally allowing duplicate links.
		return true
	}

	// If this GitHub account is already linked to the same Discord user,
	// it is not considered a conflict.
	return existingDiscordID != discordId
}

func (s *Store) Link(user User) error {
	if user.DiscordID == "" {
		return fmt.Errorf("discord id is required")
	}

	if user.GitHubUsername == "" {
		return fmt.Errorf("github username is required")
	}

	_, err := s.db.Exec(`
		INSERT INTO users (
			discord_id,
			discord_name,
			github_username,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(discord_id) DO UPDATE SET
			discord_name = excluded.discord_name,
			github_username = excluded.github_username,
			updated_at = CURRENT_TIMESTAMP;
	`, user.DiscordID, user.DiscordName, user.GitHubUsername)

	return err
}

func (s *Store) GetByDiscordID(discordID string) (User, bool, error) {
	row := s.db.QueryRow(`
		SELECT discord_id, discord_name, github_username
		FROM users
		WHERE discord_id = ?;
	`, discordID)

	var user User

	err := row.Scan(
		&user.DiscordID,
		&user.DiscordName,
		&user.GitHubUsername,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false, nil
		}

		return User{}, false, err
	}

	return user, true, nil
}

func (s *Store) AddDomain(ctx context.Context, discordID string, domain Domain) error {
	if discordID == "" {
		return fmt.Errorf("discord id is required")
	}

	if domain.Subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}

	if domain.RecordType == "" {
		return fmt.Errorf("record type is required")
	}

	if domain.Value == "" {
		return fmt.Errorf("value is required")
	}

	status := domain.Status
	if status == "" {
		status = "pending"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO domains (
			discord_id,
			subdomain,
			record_type,
			value,
			status,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(subdomain) DO UPDATE SET
			discord_id = excluded.discord_id,
			record_type = excluded.record_type,
			value = excluded.value,
			status = excluded.status,
			updated_at = CURRENT_TIMESTAMP;
	`, discordID, domain.Subdomain, domain.RecordType, domain.Value, status)

	return err
}

func (s *Store) GetDomainsByDiscordID(ctx context.Context, discordID string) ([]Domain, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT subdomain, record_type, value, status
		FROM domains
		WHERE discord_id = ?
		ORDER BY subdomain ASC;
	`, discordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain

	for rows.Next() {
		var d Domain

		if err := rows.Scan(&d.Subdomain, &d.RecordType, &d.Value, &d.Status); err != nil {
			return nil, err
		}

		domains = append(domains, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

func (s *Store) DeleteDomain(ctx context.Context, discordID string, subdomain string) error {
	if discordID == "" {
		return fmt.Errorf("discord id is required")
	}

	if subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}

	_, err := s.db.ExecContext(ctx, `
		DELETE FROM domains
		WHERE discord_id = ?
		  AND subdomain = ?;
	`, discordID, subdomain)

	return err
}

func (s *Store) UpdateDomainStatus(ctx context.Context, subdomain string, status string) error {
	if subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}

	if status == "" {
		return fmt.Errorf("status is required")
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE domains
		SET status = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE subdomain = ?;
	`, status, subdomain)

	return err
}
