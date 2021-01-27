package messaging

import (
	"time"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/gofrs/uuid"
)

const (
	// NewItemType is the metadata for messages that defines the body of message as new incoming item
	NewItemType MessageType = iota
)

// MessageType defines types of messages
//go:generate stringer -type=MessageType
type MessageType uint

// MessageBody is the useful payload of message. Do not add methods to this interface, must satisfy json.RawMessage
type MessageBody interface{}

// MessageEnvelope defines shared fields for MQ message with message type as action key, any metadata (e.g. opentracing) and Msg as actual message body content
// This is top level type in message body.
type MessageEnvelope struct {
	Type     MessageType       `json:"type,int"`
	Metadata map[string]string `json:"metadata,string"`
	Msg      MessageBody
}

//NewItemBody defines New Item message body
type NewItemBody struct {
	*entity.ItemCore
}

//NewItemMessageEnvelope creates message envelope with message type and basic item
func NewItemMessageEnvelope(
	publicationUUID uuid.UUID,
	title string,
	description string,
	content string,
	url string,
	languageCode string,
	publishedDate time.Time) (*MessageEnvelope, error) {

	itemCore := entity.NewItemCore()
	itemCore.PublicationUUID = publicationUUID
	itemCore.PublishedDate = publishedDate
	itemCore.Title = title
	itemCore.Description = description
	itemCore.Content = content
	itemCore.URL = url
	itemCore.LanguageCode = languageCode
	if err := itemCore.Validate(); err != nil {
		return &MessageEnvelope{}, err
	}
	return &MessageEnvelope{
		Type: NewItemType,
		Msg:  NewItemBody{itemCore},
	}, nil
}
