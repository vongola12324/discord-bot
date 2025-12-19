package settings

// SettingScope defines the scope of a setting
type SettingScope string

const (
	ScopeGuild SettingScope = "guild"
	ScopeUser  SettingScope = "user"
)

// SettingType defines the data type of a setting
type SettingType string

const (
	TypeString  SettingType = "string"
	TypeInt     SettingType = "int"
	TypeBool    SettingType = "bool"
	TypeSelect  SettingType = "select"
	TypeChannel SettingType = "channel"
)

// SettingDefinition defines a single setting item
type SettingDefinition struct {
	Key                string
	Module             string
	Scope              SettingScope
	Type               SettingType
	Default            interface{}
	Options            []string // Only for TypeSelect
	Validator          func(val interface{}) error
	LabelKey           string // i18n key for label
	DescKey            string // i18n key for description
	RequiredPermission int64  // Required permission for ScopeGuild
}

// Configurable is an interface for commands that have settings
type Configurable interface {
	Settings() []SettingDefinition
}
