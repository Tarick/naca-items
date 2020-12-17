package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/Tarick/naca-items/internal/graph/generated"
	uuidImpl "github.com/gofrs/uuid"
)

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

func (r *queryResolver) ItemsByPublicationUUID(ctx context.Context, publicationUUID string) ([]*entity.Item, error) {
	publUUID, err := uuidImpl.FromString(publicationUUID)
	if err != nil {
		return nil, err
	}
	return r.ItemsRepository.GetItemsByPublicationUUID(ctx, publUUID)
}

func (r *queryResolver) ItemsByPublicationUUIDSortedByPublishedUUID(ctx context.Context, publicationUUID string, orderAsc *bool) ([]*entity.Item, error) {
	publUUID, err := uuidImpl.FromString(publicationUUID)
	if err != nil {
		return nil, err
	}
	sortOrder := false
	if orderAsc != nil && *orderAsc == true {
		sortOrder = true

	}
	return r.ItemsRepository.GetItemsByPublicationUUIDSortByPublishedDate(ctx, publUUID, sortOrder)
}

// Item returns generated.ItemResolver implementation.
func (r *Resolver) Item() generated.ItemResolver { return &itemResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type itemResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
