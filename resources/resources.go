package resources

import "embed"

// I18n contains all embedded i18n translation files
//
//go:embed i18n/*.json
var I18n embed.FS

// I18nBasePath is the base path for i18n resources within the embedded filesystem
const I18nBasePath = "i18n"
