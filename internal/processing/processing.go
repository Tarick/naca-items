package processing

import (
	"encoding/json"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/Tarick/naca-items/internal/messaging"
)

// Logger interface
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// ItemsRepository defines repository methods
type ItemsRepository interface {
	Create(*entity.Item) error
	ItemExists(*entity.Item) (bool, error)
}

// processor is container for business logic
type processor struct {
	repository ItemsRepository
	logger     Logger
}

// New creates processor for messaging feeds operations
func New(repository ItemsRepository, logger Logger) *processor {
	return &processor{
		repository,
		logger,
	}
}

// Process is a gateway for message consumption - handles incoming data and calls related handlers
// It uses json.RawMessage to delay the unmarshalling of message content - Type is unmarshalled first.
func (p *processor) Process(data []byte) error {
	var msg json.RawMessage
	message := messaging.MessageEnvelope{Msg: &msg}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	switch message.Type {
	case messaging.NewItem:
		var msgBody messaging.NewItemBody
		if err := json.Unmarshal(msg, &msgBody); err != nil {
			p.logger.Error("Failure unmarshalling NewItem body content: ", err)
			return err
		}
		item := entity.Item{
			UUID:            msgBody.UUID,
			PublicationUUID: msgBody.PublicationUUID,
			Description:     msgBody.Description,
			Content:         msgBody.Content,
			Source:          msgBody.Source,
			Author:          msgBody.Author,
			LanguageCode:    msgBody.LanguageCode,
		}
		return p.CreateItem(item)
	default:
		p.logger.Error("Undefined message type: ", message.Type)
		// TODO: implement common errors
		return fmt.Errorf("Undefined message type: %v", message.Type)
	}
}

// CreateItem adds it to the system
func (p *processor) CreateItem(item entity.Item) error {
	itemExist, err := p.repository.ItemExists(&item)
	if err != nil {
		return fmt.Errorf("couldn't get item from repository, %w", err)
	}
	if itemExist {
		errorMsg := fmt.Errorf("item %s already exist in repository", item.UUID)
		p.logger.Error(errorMsg)
		return errorMsg
	}
	if err := p.repository.Create(&item); err != nil {
		return err
	}
	p.logger.Info("Processed new item ", item.UUID, ", publication ", item.PublicationUUID)
	return nil
}
