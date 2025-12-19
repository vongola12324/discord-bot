package store

import (
	"database/sql"
	"hiei-discord-bot/internal/models"
	"hiei-discord-bot/internal/settings"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Create table
	var query string
	// 1. settings
	query = `
	CREATE TABLE IF NOT EXISTS settings (
		scope TEXT NOT NULL,
		target_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY (scope, target_id, key)
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}
	// 2. local_command_versions
	query = `
	CREATE TABLE IF NOT EXISTS local_command_versions (
		command_name TEXT NOT NULL,
		version TEXT NOT NULL,
		build_time TEXT NOT NULL,
		PRIMARY KEY (command_name)
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}
	// 3. guild_command_versions
	query = `
	CREATE TABLE IF NOT EXISTS guild_command_versions (
		guild_id TEXT NOT NULL,
		command_name TEXT NOT NULL,
		version_time TEXT NOT NULL,
		PRIMARY KEY (guild_id, command_name)
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) GetSetting(scope settings.SettingScope, targetID, key string) (string, error) {
	var value string
	query := "SELECT value FROM settings WHERE scope = ? AND target_id = ? AND key = ?"
	err := s.db.QueryRow(query, string(scope), targetID, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *SQLiteStore) SetSetting(scope settings.SettingScope, targetID, key, value string) error {
	query := `
	INSERT INTO settings (scope, target_id, key, value)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(scope, target_id, key) DO UPDATE SET value = excluded.value;`
	_, err := s.db.Exec(query, string(scope), targetID, key, value)
	return err
}

func (s *SQLiteStore) GetLocalCommandVersionAndBuildTime(command_name string) (models.CommandVersion, error) {
	var version string
	var build_time string
	query := "SELECT version, build_time FROM local_command_versions WHERE command_name = ?"
	err := s.db.QueryRow(query, command_name).Scan(&version, &build_time)
	if err == sql.ErrNoRows {
		return models.CommandVersion{}, nil
	}
	build_time_tmp, _ := time.Parse(time.RFC3339, build_time)
	return models.CommandVersion{
		Version:   version,
		BuildTime: build_time_tmp,
	}, err
}

func (s *SQLiteStore) UpdateLocalCommandVersionAndBuildTime(command_name string, command_version models.CommandVersion) error {
	query := `
	INSERT INTO local_command_versions (command_name, version, build_time)
	VALUES (?, ?, ?)
	ON CONFLICT(command_name)
	DO UPDATE SET
		version = excluded.version,
		build_time = excluded.build_time;
	`
	_, err := s.db.Exec(query, command_name, command_version.Version, command_version.BuildTime.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) GetGuildCommandLastVersionTime(guild_id string, command_name string) (*time.Time, error) {
	var version_time string
	query := "SELECT version_time FROM guild_command_versions WHERE guild_id = ? AND command_name = ?"
	err := s.db.QueryRow(query, guild_id, command_name).Scan(&version_time)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	version_time_tmp, _ := time.Parse(time.RFC3339, version_time)
	return &version_time_tmp, err
}

func (s *SQLiteStore) UpdateGuildCommandLastVersionTime(guild_id string, command_name string, upload_time time.Time) error {
	query := `
	INSERT INTO guild_command_versions (guild_id, command_name, version_time)
	VALUES (?, ?, ?)
	ON CONFLICT(guild_id, command_name)
	DO UPDATE SET 
		version_time = excluded.version_time
	`
	_, err := s.db.Exec(query, guild_id, command_name, upload_time.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) GetAllLocalCommandNames() ([]string, error) {
	query := "SELECT command_name FROM local_command_versions"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (s *SQLiteStore) DeleteLocalCommandVersion(command_name string) error {
	query := "DELETE FROM local_command_versions WHERE command_name = ?"
	_, err := s.db.Exec(query, command_name)
	return err
}

func (s *SQLiteStore) DeleteGuildCommandVersion(guild_id string, command_name string) error {
	query := "DELETE FROM guild_command_versions WHERE guild_id = ? AND command_name = ?"
	_, err := s.db.Exec(query, guild_id, command_name)
	return err
}
