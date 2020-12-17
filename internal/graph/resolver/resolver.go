package resolver

import (
	"context"

	"github.com/Tarick/naca-items/internal/entity"
	// Rename to uuidImpl since uuid is masked in functions - used with gqlgen code generation
	uuidImpl "github.com/gofrs/uuid"
)

//go:generate go run github.com/99designs/gqlgen

// Resolver uses dependency injection
type Resolver struct {
	ItemsRepository ItemsRepository
}

// ItemsRepository is the interface for repository implementation
type ItemsRepository interface {
	GetItemByUUID(context.Context, uuidImpl.UUID) (*entity.Item, error)
	GetItems(context.Context) ([]*entity.Item, error)
	GetItemsByPublicationUUID(context.Context, uuidImpl.UUID) ([]*entity.Item, error)
	GetItemsByPublicationUUIDSortByPublishedDate(context.Context, uuidImpl.UUID, bool) ([]*entity.Item, error)
}
