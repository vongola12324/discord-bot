package settings

import (
	"fmt"
	"hiei-discord-bot/internal/models"
	"sync"
	"time"
)

// Manager handles setting registration and access
type Manager struct {
	definitions map[string]SettingDefinition
	store       Store
	mu          sync.RWMutex
}

// Store interface for persistence
type Store interface {
	GetSetting(scope SettingScope, targetID, key string) (string, error)
	SetSetting(scope SettingScope, targetID, key, value string) error
	GetLocalCommandVersionAndBuildTime(command_name string) (models.CommandVersion, error)
	UpdateLocalCommandVersionAndBuildTime(command_name string, command_version models.CommandVersion) error
	GetGuildCommandLastVersionTime(guild_id string, command_name string) (*time.Time, error)
	UpdateGuildCommandLastVersionTime(guild_id string, command_name string, upload_time time.Time) error
	GetAllLocalCommandNames() ([]string, error)
	DeleteLocalCommandVersion(command_name string) error
	DeleteGuildCommandVersion(guild_id string, command_name string) error
}

var instance *Manager
var once sync.Once

// GetManager returns the singleton manager instance
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			definitions: make(map[string]SettingDefinition),
		}
	})
	return instance
}

// SetStore sets the storage engine
func (mgr *Manager) SetStore(s Store) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.store = s
}

// Register adds a setting definition
func (mgr *Manager) Register(def SettingDefinition) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.definitions[def.Key] = def
}

// GetDefinitions returns all registered definitions
func (mgr *Manager) GetDefinitions() []SettingDefinition {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	defs := make([]SettingDefinition, 0, len(mgr.definitions))
	for _, def := range mgr.definitions {
		defs = append(defs, def)
	}
	return defs
}

// GetSettingValue retrieves a setting value, falling back to default if not set
func (mgr *Manager) GetSettingValue(scope SettingScope, targetID, key string) (interface{}, error) {
	mgr.mu.RLock()
	def, exists := mgr.definitions[key]
	mgr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("setting %s not found", key)
	}

	if mgr.store == nil {
		return def.Default, nil
	}

	valStr, err := mgr.store.GetSetting(scope, targetID, key)
	if err != nil || valStr == "" {
		return def.Default, nil
	}

	// Convert string value back to original type
	return mgr.convertSettingValue(valStr, def.Type)
}

// SetSettingValue validates and stores a setting value
func (mgr *Manager) SetSettingValue(scope SettingScope, targetID, key string, value interface{}) error {
	mgr.mu.RLock()
	def, exists := mgr.definitions[key]
	mgr.mu.RUnlock()

	if !exists {
		return fmt.Errorf("setting %s not found", key)
	}

	// Validate
	if def.Validator != nil {
		if err := def.Validator(value); err != nil {
			return err
		}
	}

	if mgr.store == nil {
		return fmt.Errorf("store not initialized")
	}

	return mgr.store.SetSetting(scope, targetID, key, fmt.Sprintf("%v", value))
}

func (mgr *Manager) GetLocalCommandVersionAndBuildTime(commandName string) (models.CommandVersion, error) {
	if mgr.store == nil {
		return models.CommandVersion{}, fmt.Errorf("store not initialized")
	}
	return mgr.store.GetLocalCommandVersionAndBuildTime(commandName)
}

func (mgr *Manager) UpdateLocalCommandVersionAndBuildTime(commandName string, version models.CommandVersion) error {
	if mgr.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return mgr.store.UpdateLocalCommandVersionAndBuildTime(commandName, version)
}

func (mgr *Manager) GetGuildCommandLastVersionTime(guildID string, commandName string) (*time.Time, error) {
	if mgr.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return mgr.store.GetGuildCommandLastVersionTime(guildID, commandName)
}

func (mgr *Manager) UpdateGuildCommandLastVersionTime(guildID string, commandName string, uploadTime time.Time) error {
	if mgr.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return mgr.store.UpdateGuildCommandLastVersionTime(guildID, commandName, uploadTime)
}

func (mgr *Manager) GetAllLocalCommandNames() ([]string, error) {
	if mgr.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return mgr.store.GetAllLocalCommandNames()
}

func (mgr *Manager) DeleteLocalCommandVersion(commandName string) error {
	if mgr.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return mgr.store.DeleteLocalCommandVersion(commandName)
}

func (mgr *Manager) DeleteGuildCommandVersion(guildID string, commandName string) error {
	if mgr.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return mgr.store.DeleteGuildCommandVersion(guildID, commandName)
}

func (mgr *Manager) convertSettingValue(val string, t SettingType) (interface{}, error) {
	switch t {
	case TypeInt:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err
	case TypeBool:
		return val == "true", nil
	default:
		return val, nil
	}
}
