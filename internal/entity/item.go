package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
)

// Item defines news item type
type Item struct {
	UUID uuid.UUID `json:"uuid"`
	*ItemCore
}

// Validate checks validity of item fields
func (item *Item) Validate() error {
	return validation.ValidateStruct(&item,
		validation.Field(&item.UUID, validation.Required, is.UUID, validation.By(checkUUIDNotNil)),
	)
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
	return validation.ValidateStruct(&core,
		validation.Field(&core.PublicationUUID, validation.Required, is.UUID, validation.By(checkUUIDNotNil)),
		validation.Field(&core.PublishedDate, validation.Required),
		validation.Field(&core.Title, validation.Required, validation.Length(5, 400)),
		validation.Field(&core.Description, validation.Required, validation.Length(5, 0)),
		validation.Field(&core.Content, validation.Required, validation.Length(5, 0)),
		validation.Field(&core.Source, validation.Required, validation.Length(5, 100), is.URL),
		validation.Field(&core.Author, validation.Required, validation.Length(3, 100)),
		validation.Field(&core.LanguageCode, validation.Required, validation.Length(2, 2), isLanguageCode),
	)
}

// validation helper to check UUID
func checkUUIDNotNil(value interface{}) error {
	u, _ := value.(uuid.UUID)
	if u == uuid.Nil {
		return errors.New("uuid is nil")
	}
	return nil
}

//NewItemCore abstracts ItemCore creation
func NewItemCore() *ItemCore {
	return &ItemCore{}
}

var isLanguageCode = validation.NewStringRuleWithError(
	govalidator.IsISO693Alpha2,
	validation.NewError("validation_is_language_code_2_letter", "must be a valid two-letter ISO693Alpha2 language code"))

// NewItem creates new item with set UUID v5, using PublicationUUID as a namespace and Title and PublishedDate as a key
// This ensures uniquness of published item
func NewItem(core *ItemCore) *Item {
	item := &Item{}
	item.ItemCore = core
	item.UUID = uuid.NewV5(item.PublicationUUID, fmt.Sprint(item.Title, "_", item.PublishedDate))
	return item
}
