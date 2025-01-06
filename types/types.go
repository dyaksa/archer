package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NullUuid struct {
	UUID  uuid.UUID
	Valid bool
}

func (nu *NullUuid) Scan(value interface{}) error {
	if value == nil {
		nu.UUID = uuid.UUID{}
		nu.Valid = false
		return nil
	}

	var u uuid.UUID
	err := u.Scan(value)
	if err != nil {
		return err
	}

	nu.UUID = u
	nu.Valid = true
	return nil
}

// MarshalJSON for NullUuid
func (nu *NullUuid) MarshalJSON() ([]byte, error) {
	if !nu.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nu.UUID)
}

// UnmarshalJSON for NullUuid
func (nu *NullUuid) UnmarshalJSON(b []byte) error {
	nu.Valid = string(b) != "null"
	e := json.Unmarshal(b, &nu.UUID)
	return e
}

type NullString struct {
	sql.NullString
}

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON for NullString
func (ns *NullString) UnmarshalJSON(b []byte) error {
	ns.Valid = string(b) != "null"
	e := json.Unmarshal(b, &ns.String)
	return e
}

type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON for NullInt64
func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

// UnmarshalJSON for NullInt64
func (ns *NullInt64) UnmarshalJSON(b []byte) error {
	ns.Valid = string(b) != "null"
	e := json.Unmarshal(b, &ns.Int64)
	return e
}

type NullBool struct {
	sql.NullBool
}

// MarshalJSON for NullBool
func (nb *NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nb.Bool)
}

// UnmarshalJSON for NullBool
func (ns *NullBool) UnmarshalJSON(b []byte) error {
	ns.Valid = string(b) != "null"
	e := json.Unmarshal(b, &ns.Bool)
	return e
}

type NullFloat64 struct {
	sql.NullFloat64
}

// MarshalJSON for NullFloat64
func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}

// UnmarshalJSON for NullFloat64
func (ns *NullFloat64) UnmarshalJSON(b []byte) error {
	ns.Valid = string(b) != "null"
	e := json.Unmarshal(b, &ns.Float64)
	return e
}

type NullTime struct {
	sql.NullTime
}

// MarshalJSON for NullTime
func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
	return []byte(val), nil
}

// UnmarshalJSON for NullTime
func (ns *NullTime) UnmarshalJSON(b []byte) error {
	ns.Valid = string(b) != "null"
	e := json.Unmarshal(b, &ns.Time)
	return e
}
