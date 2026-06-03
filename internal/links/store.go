package users

import (
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
		CREATE TABLE IF NOT EXISTS users (
			discord_id TEXT PRIMARY KEY,
			discord_name TEXT NOT NULL,
			github_username TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
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
