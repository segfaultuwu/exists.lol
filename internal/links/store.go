package links

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	path string
	mu   sync.Mutex
}

type Link struct {
	DiscordID      string `json:"discord_id"`
	DiscordName    string `json:"discord_name"`
	GitHubUsername string `json:"github_username"`
}

func NewStore(path string) *Store {
	return &Store{
		path: path,
	}
}

func (s *Store) Set(link Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	all, err := s.load()
	if err != nil {
		return err
	}

	all[link.DiscordID] = link

	return s.save(all)
}

func (s *Store) Get(discordID string) (Link, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	all, err := s.load()
	if err != nil {
		return Link{}, false, err
	}

	link, ok := all[discordID]
	return link, ok, nil
}

func (s *Store) load() (map[string]Link, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]Link{}, nil
		}

		return nil, err
	}

	if len(raw) == 0 {
		return map[string]Link{}, nil
	}

	var links map[string]Link
	if err := json.Unmarshal(raw, &links); err != nil {
		return nil, err
	}

	return links, nil
}

func (s *Store) save(links map[string]Link) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		return err
	}

	raw = append(raw, '\n')

	return os.WriteFile(s.path, raw, 0600)
}
