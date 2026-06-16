package models

import (
	"database/sql/driver"
	"fmt"
)

type TrackStatus int

const (
	TrackStatusQueued TrackStatus = iota
	TrackStatusPlaying
	TrackStatusSkipped
	TrackStatusPlayed
	TrackStatusProposed
)

type MusicService int

const (
	ServiceSpotify MusicService = iota
	ServiceTidal
	ServiceSoundcloud
)

var musicServiceName = map[MusicService]string{
	ServiceSpotify:    "Spotify",
	ServiceTidal:      "Tidal",
	ServiceSoundcloud: "Soundcloud",
}

func (m MusicService) String() string {
	return musicServiceName[m]
}

var trackStatusName = map[TrackStatus]string{
	TrackStatusQueued:   "queued",
	TrackStatusPlaying:  "playing",
	TrackStatusSkipped:  "skipped",
	TrackStatusPlayed:   "played",
	TrackStatusProposed: "proposed",
}

func (s TrackStatus) String() string {
	return trackStatusName[s]
}

func (s *TrackStatus) Scan(value any) error {
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		return fmt.Errorf("TrackStatus must be a string or []byte")
	}

	switch strVal {
	case "queued":
		*s = TrackStatusQueued
	case "playing":
		*s = TrackStatusPlaying
	case "skipped":
		*s = TrackStatusSkipped
	case "played":
		*s = TrackStatusPlayed
	case "proposed":
		*s = TrackStatusProposed
	default:
		return fmt.Errorf("unknown track status: %s", strVal)
	}
	return nil
}

func (s TrackStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s TrackStatus) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%q", s.String()), nil
}

type UserRole int

const (
	RoleGuest UserRole = iota
	RoleDJ
	RoleHost
)

var userRoleName = map[UserRole]string{
	RoleGuest: "guest",
	RoleDJ:    "dj",
	RoleHost:  "host",
}

func (r UserRole) String() string {
	return userRoleName[r]
}

func (r UserRole) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%q", r.String()), nil
}

func (r UserRole) Value() (driver.Value, error) {
	return r.String(), nil
}

func (r *UserRole) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("UserRole must be a string or []byte")
	}
	switch s {
	case "Guest":
		*r = RoleGuest
	case "DJ":
		*r = RoleDJ
	case "Host":
		*r = RoleHost
	default:
		return fmt.Errorf("invalid UserRole: %s", s)
	}
	return nil
}

type UserType int

const (
	UserTypeRegistered UserType = iota // 0
	UserTypeGuest                      // 1
)

func (t UserType) String() string {
	return [...]string{"Registered", "Guest"}[t]
}

func (t UserType) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%q", t.String()), nil
}

func (t UserType) Value() (driver.Value, error) {
	return t.String(), nil
}

func (t *UserType) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("UserType must be a string or []byte")
	}
	switch s {
	case "Registered":
		*t = UserTypeRegistered
	case "Guest":
		*t = UserTypeGuest
	default:
		return fmt.Errorf("invalid UserType: %s", s)
	}
	return nil
}
