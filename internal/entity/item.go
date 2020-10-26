package entity

import (
	"fmt"

	"github.com/gofrs/uuid"
)

// Item defines news item type
type Item struct {
	UUID            uuid.UUID `json:"uuid"`
	PublicationUUID uuid.UUID `json:"publication_uuid"`
	Description     string    `json:"description"`
	Content         string    `json:"content"`
	Source          string    `json:"source"`
	Author          string    `json:"author"`
	LanguageCode    string    `json:"language_code"`
}

func (item *Item) String() string {
	return fmt.Sprintf("PublicationUUID: %v, Source: %s", item.PublicationUUID, item.Source)
}
