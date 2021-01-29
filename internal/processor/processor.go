package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otLog "github.com/opentracing/opentracing-go/log"
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
	tracer     opentracing.Tracer
}

// New creates processor for messaging feeds operations
func New(repository ItemsRepository, logger Logger, tracer opentracing.Tracer) *processor {
	return &processor{
		repository,
		logger,
		tracer,
	}
}

// Process is a gateway for message consumption - handles incoming data and calls related handlers
// It uses json.RawMessage to delay the unmarshalling of message content - Type is unmarshalled first to figure out what type of message it is.
func (p *processor) Process(data []byte) error {
	var msg json.RawMessage
	message := MessageEnvelope{Msg: &msg}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	// Setup tracing span
	messageSpanContext, err := p.tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(message.Metadata))
	if err != nil {
		p.logger.Debug("No tracing information in message metadata: ", err)
	}
	span := p.tracer.StartSpan("process-message", opentracing.FollowsFrom(messageSpanContext))
	defer span.Finish()
	ext.Component.Set(span, "ItemsProcessor")
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	switch message.Type {
	case NewItemType:
		var msgBody NewItemBody
		if err := json.Unmarshal(msg, &msgBody); err != nil {
			p.logger.Error("Failure unmarshalling body content: ", err)
			return err
		}
		if err := msgBody.ItemCore.Validate(); err != nil {
			return err
		}
		return p.ProcessNewItem(ctx, msgBody.ItemCore)
	default:
		p.logger.Error("Undefined message type: ", message.Type)
		// TODO: implement common errors
		return fmt.Errorf("Undefined message type: %v", message.Type)
	}
}

func (p *processor) ProcessNewItem(ctx context.Context, itemCore *entity.ItemCore) error {
	item := entity.NewFilledItem(itemCore)
	//TODO: next process steps are here
	return p.CreateItem(ctx, item)
}

// CreateItem adds it to the system
func (p *processor) CreateItem(ctx context.Context, item *entity.Item) error {
	span, ctx := p.setupTracingSpan(ctx, "create-new-item")
	defer span.Finish()
	span.SetTag("item.UUID", item.UUID)
	span.SetTag("item.publicationUUID", item.PublicationUUID)
	itemExist, err := p.repository.ItemExists(ctx, item)
	if err != nil {
		checkErrMsg := fmt.Errorf("couldn't check if item with UUID %s exists in repository", item.UUID)
		span.LogFields(
			otLog.Error(checkErrMsg),
		)
		return checkErrMsg
	}
	if itemExist {
		existsErrMsg := fmt.Errorf("item %s already exist in repository", item.UUID)
		span.LogFields(
			otLog.Error(existsErrMsg),
		)
		p.logger.Error(existsErrMsg)
		return existsErrMsg
	}
	if err := p.repository.Create(ctx, item); err != nil {
		span.LogFields(
			otLog.Error(err),
		)
		return err
	}
	p.logger.Info("Processed new item ", item.UUID, ", publication ", item.PublicationUUID)
	span.LogKV("event", "created item")
	return nil
}

func (p *processor) setupTracingSpan(ctx context.Context, name string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, p.tracer, name)
	ext.Component.Set(span, "ItemsProcessor")
	return span, ctx
}
