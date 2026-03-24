package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type ResourceLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type ResourceLinks []ResourceLink

func (rl *ResourceLinks) Scan(value interface{}) error {
	if value == nil {
		*rl = ResourceLinks{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, rl)
}

func (rl ResourceLinks) Value() (driver.Value, error) {
	if len(rl) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(rl)
}

type ResourcePage struct {
	Slug      string        `db:"slug"       json:"slug"`
	Title     string        `db:"title"      json:"title"`
	Content   string        `db:"content"    json:"content"`
	Links     ResourceLinks `db:"links"      json:"links"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt time.Time     `db:"updated_at" json:"updated_at"`
}

type ResourcePageUpdateRequest struct {
	Title   *string         `json:"title,omitempty"`
	Content *string         `json:"content,omitempty"`
	Links   *[]ResourceLink `json:"links,omitempty"`
}
