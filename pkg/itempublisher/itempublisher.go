// Package itempublisher provides message publishing client for items
package itempublisher

import (
	"encoding/json"
	"time"

	"github.com/Tarick/naca-items/internal/messaging/nsqclient/producer"
	"github.com/Tarick/naca-items/internal/processor"
	"github.com/gofrs/uuid"
)

type messageProducer interface {
	Publish([]byte) error
}

// TODO: add logger
type messagePublisher struct {
	messageProducer messageProducer
}

func (p *messagePublisher) PublishNewItem(
	metadata map[string]string,
	publicationUUID uuid.UUID,
	title string,
	description string,
	content string,
	url string,
	languageCode string,
	publishedDate time.Time) error {

	message, err := processor.NewItemMessageEnvelope(metadata, publicationUUID, title, description, content, url, languageCode, publishedDate)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return p.messageProducer.Publish(bytes)
}

// New creates message publisher
func New(host string, topic string) (*messagePublisher, error) {
	producer, err := producer.New(&producer.MessageProducerConfig{Host: host, Topic: topic})
	if err != nil {
		return nil, err
	}
	return &messagePublisher{messageProducer: producer}, nil
}
