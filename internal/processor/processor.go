package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/Tarick/naca-items/internal/messaging"
)

// ItemsRepository defines repository methods
type ItemsRepository interface {
	Create(context.Context, *entity.Item) error
	ItemExists(context.Context, *entity.Item) (bool, error)
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
// It uses json.RawMessage to delay the unmarshalling of message content - Type is unmarshalled first to figure out what type of message it is.
func (p *processor) Process(data []byte) error {
	var msg json.RawMessage
	message := messaging.MessageEnvelope{Msg: &msg}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	switch message.Type {
	case messaging.NewItemType:
		var msgBody messaging.NewItemBody
		if err := json.Unmarshal(msg, &msgBody); err != nil {
			p.logger.Error("Failure unmarshalling body content: ", err)
			return err
		}
		if err := msgBody.ItemCore.Validate(); err != nil {
			return err
		}
		return p.ProcessNewItem(msgBody.ItemCore)
	default:
		p.logger.Error("Undefined message type: ", message.Type)
		// TODO: implement common errors
		return fmt.Errorf("Undefined message type: %v", message.Type)
	}
}

func (p *processor) ProcessNewItem(itemCore *entity.ItemCore) error {
	item := entity.NewItem(itemCore)
	//TODO: next process steps are here
	return p.CreateItem(item)
}

// CreateItem adds it to the system
func (p *processor) CreateItem(item *entity.Item) error {
	itemExist, err := p.repository.ItemExists(context.Background(), item)
	if err != nil {
		return fmt.Errorf("couldn't get item from repository, %w", err)
	}
	if itemExist {
		errorMsg := fmt.Errorf("item %s already exist in repository", item.UUID)
		p.logger.Error(errorMsg)
		return errorMsg
	}
	if err := p.repository.Create(context.Background(), item); err != nil {
		return err
	}
	p.logger.Info("Processed new item ", item.UUID, ", publication ", item.PublicationUUID)
	return nil
}
