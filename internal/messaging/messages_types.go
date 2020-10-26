package messaging

import (
	"github.com/Tarick/naca-items/internal/entity"
)

const (
	NewItem MessageType = iota
)

// MessageType defines types of messages
//go:generate stringer -type=MessageType
type MessageType uint

type MessageBody interface{}

// MessageEnvelope defines shared fields for MQ message with message type as action key and Msg as actual message body content
// This is top level type in message body.
type MessageEnvelope struct {
	Type MessageType `json:"type,int"`
	Msg  MessageBody
}

type NewItemBody struct {
	*entity.Item
}

func NewItemMessage() *MessageEnvelope {
	return &MessageEnvelope{
		Type: NewItem,
		Msg:  NewItemBody{},
	}
}
