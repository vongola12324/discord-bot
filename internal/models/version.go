package models

import "time"

type CommandVersion struct {
	Version   string
	BuildTime time.Time
}
