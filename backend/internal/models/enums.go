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
)

var trackStatusName = map[TrackStatus]string{
	TrackStatusQueued:  "queued",
	TrackStatusPlaying: "playing",
	TrackStatusSkipped: "skipped",
	TrackStatusPlayed:  "played",
}

func (s TrackStatus) String() string {
	return trackStatusName[s]
}

func (s *TrackStatus) Scan(value any) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("TrackStatus must be a string")
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
	default:
		return fmt.Errorf("unknown track status: %s", strVal)
	}
	return nil
}

func (s TrackStatus) Value() (any, error) {
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
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("UserRole must be a string")
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
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("UserType must be a string")
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
