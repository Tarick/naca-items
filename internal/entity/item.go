package entity

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

// Item defines news item type
type Item struct {
	UUID uuid.UUID `json:"uuid"`
	*ItemCore
}

func (item *Item) String() string {
	return fmt.Sprintf("PublicationUUID: %v, Source: %s", item.PublicationUUID, item.Source)
}

// ItemCore defines essential for creation fields of Item
// Any submitter of item must provide those.
type ItemCore struct {
	PublicationUUID uuid.UUID `json:"publication_uuid"`
	PublishedDate   time.Time `json:"publication_date"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Content         string    `json:"content"`
	Source          string    `json:"source"`
	Author          string    `json:"author"`
	LanguageCode    string    `json:"language_code"`
}

// Validate checks core item fields
func (core *ItemCore) Validate() error {
	//TODO: do validation
	return nil
}

//NewItemCore abstracts ItemCore creation
func NewItemCore() *ItemCore {
	return &ItemCore{}
}

// NewItem creates new item with set UUID v5, using PublicationUUID as a namespace and Title and PublishedDate as a key
// This ensures uniquness of published item
func NewItem(core *ItemCore) *Item {
	item := &Item{}
	item.ItemCore = core
	item.UUID = uuid.NewV5(item.PublicationUUID, fmt.Sprint(item.Title, "_", item.PublishedDate))
	return item
}
