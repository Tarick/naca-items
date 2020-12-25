package resolver

import (
	"context"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"
	// Rename to uuidImpl since uuid is masked in functions - used with gqlgen code generation
	"github.com/gofrs/uuid"
	uuidImpl "github.com/gofrs/uuid"
)

//go:generate go run github.com/99designs/gqlgen

// Resolver uses dependency injection
type Resolver struct {
	ItemsRepository ItemsRepository
}

// ItemsRepository is the interface for repository implementation
type ItemsRepository interface {
	GetItemByUUID(context.Context, uuid.UUID) (*entity.Item, error)
	GetItems(context.Context) ([]*entity.Item, error)
	GetItemsByPublicationUUID(context.Context, uuidImpl.UUID) ([]*entity.Item, error)
	GetItemsByPublicationUUIDSortByPublishedDate(context.Context, uuidImpl.UUID, bool) ([]*entity.Item, error)
	// Needed to healthcheck
	Healthcheck(context.Context) error
}

func getItemIndexByUUID(arr []*entity.Item, id uuidImpl.UUID) (int, error) {
	for i := range arr {
		if arr[i].UUID == id {
			return i, nil
		}
	}
	return 0, fmt.Errorf("didn't find element %s in array", id)
}
