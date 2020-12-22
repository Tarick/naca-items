package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/Tarick/naca-items/internal/graph/generated"
	"github.com/Tarick/naca-items/internal/graph/model"
	uuidImpl "github.com/gofrs/uuid"
)

func (r *itemResolver) UUID(ctx context.Context, obj *entity.Item) (string, error) {
	return obj.UUID.String(), nil
}

func (r *itemResolver) PublicationUUID(ctx context.Context, obj *entity.Item) (string, error) {
	return obj.PublicationUUID.String(), nil
}

func (r *itemsConnectionResolver) Edges(ctx context.Context, obj *model.ItemsConnection) ([]*model.ItemsEdge, error) {
	edges := make([]*model.ItemsEdge, obj.ToIndex-obj.FromIndex+1)
	for i := range edges {
		edges[i] = &model.ItemsEdge{
			Node:   obj.Items[obj.FromIndex+i],
			Cursor: model.EncodeCursor(obj.Items[obj.FromIndex+i].UUID),
		}
	}
	return edges, nil
}

func (r *queryResolver) Items(ctx context.Context, publicationUUID *string, orderAsc *bool) ([]*entity.Item, error) {
	if publicationUUID == nil {
		return r.ItemsRepository.GetItems(ctx)
	} else {
		publUUID, err := uuidImpl.FromString(*publicationUUID)
		if err != nil {
			return nil, err
		}
		return r.ItemsRepository.GetItemsByPublicationUUIDSortByPublishedDate(ctx, publUUID, *orderAsc)
	}
}

func (r *queryResolver) ItemsConnection(ctx context.Context, publicationUUID *string, orderAsc *bool, first *int, after *string, last *int, before *string) (*model.ItemsConnection, error) {
	publUUID, err := uuidImpl.FromString(*publicationUUID)
	if err != nil {
		return nil, err
	}
	fetchedItems, err := r.ItemsRepository.GetItemsByPublicationUUIDSortByPublishedDate(ctx, publUUID, *orderAsc)
	if err != nil {
		return nil, err
	}
	if len(fetchedItems) == 0 {
		return &model.ItemsConnection{
			Items:     fetchedItems,
			FromIndex: 0,
			ToIndex:   0,
		}, nil
	}
	// Default slicing - the whole slice
	fromIndex := 0
	// Last item UUID
	toIndex := len(fetchedItems) - 1
	if after != nil {
		// from := fetchedItems[fromIndex].UUID
		from, err := model.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}
		fromIndex, err = getItemIndexByUUID(fetchedItems, from)
		if err != nil {
			return nil, err
		}
		// with 'after', we return the next element
		fromIndex++
	}
	if before != nil {
		to, err := model.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}
		toIndex, err = getItemIndexByUUID(fetchedItems, to)
		if err != nil {
			return nil, err
		}
		// with 'before', we return the previous elements
		toIndex--
		if fromIndex > toIndex {
			return nil, fmt.Errorf("'before' and 'after' cursors yielded negative slice of elements, incorrect cursors or changed order assumed")
		}
	}
	if first != nil {
		if *first < 0 {
			return nil, fmt.Errorf("'first' parameter cannot be negative")
		}
		toFirstIndex := fromIndex + *first - 1
		if toFirstIndex < toIndex {
			toIndex = toFirstIndex
		}
	}
	if last != nil {
		if *last < 0 {
			return nil, fmt.Errorf("'last' parameter cannot be negative")
		}
		fromLastIndex := toIndex - *last + 1
		if fromLastIndex > fromIndex {
			fromIndex = fromLastIndex
		}
	}

	return &model.ItemsConnection{
		Items:     fetchedItems,
		FromIndex: fromIndex,
		ToIndex:   toIndex,
	}, nil
}

func (r *queryResolver) Item(ctx context.Context, uuid string) (*entity.Item, error) {
	UUID, err := uuidImpl.FromString(uuid)
	if err != nil {
		return nil, err
	}
	return r.ItemsRepository.GetItemByUUID(ctx, UUID)
}

// Item returns generated.ItemResolver implementation.
func (r *Resolver) Item() generated.ItemResolver { return &itemResolver{r} }

// ItemsConnection returns generated.ItemsConnectionResolver implementation.
func (r *Resolver) ItemsConnection() generated.ItemsConnectionResolver {
	return &itemsConnectionResolver{r}
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type itemResolver struct{ *Resolver }
type itemsConnectionResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
