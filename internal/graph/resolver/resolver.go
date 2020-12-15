package resolver

//go:generate go run github.com/99designs/gqlgen

import (
	"context"

	"github.com/Tarick/naca-items/internal/entity"
	uuidImpl "github.com/gofrs/uuid"
)

// ItemsRepository defines repository methods
type ItemsRepository interface {
	GetItemByUUID(ctx context.Context, UUID uuidImpl.UUID) (*entity.Item, error)
	GetItems(ctx context.Context) ([]*entity.Item, error)
}

// Resolver uses dependency injection
type Resolver struct {
	ItemsRepository ItemsRepository
}

func (r *itemResolver) UUID(ctx context.Context, obj *entity.Item) (string, error) {
	return obj.UUID.String(), nil
}

func (r *itemResolver) PublicationUUID(ctx context.Context, obj *entity.Item) (string, error) {
	return obj.PublicationUUID.String(), nil
}

func (r *queryResolver) Items(ctx context.Context) ([]*entity.Item, error) {
	return r.ItemsRepository.GetItems(ctx)
}

func (r *queryResolver) Item(ctx context.Context, uuid string) (*entity.Item, error) {
	UUID, err := uuidImpl.FromString(uuid)
	if err != nil {
		return nil, err
	}
	return r.ItemsRepository.GetItemByUUID(ctx, UUID)
}
