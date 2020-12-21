package model

import (
	"encoding/base64"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/gofrs/uuid"
)

type ItemsConnection struct {
	Items     []*entity.Item
	From      uuid.UUID
	To        uuid.UUID
	FromIndex int
	ToIndex   int
}

func (i *ItemsConnection) TotalCount() int {
	return len(i.Items)
}

func (i *ItemsConnection) PageInfo() PageInfo {
	start := EncodeCursor(i.From)
	to := EncodeCursor(i.To)
	return PageInfo{
		StartCursor:     &start,
		EndCursor:       &to,
		HasNextPage:     i.ToIndex < len(i.Items)-1,
		HasPreviousPage: i.FromIndex > 0,
	}
}

func EncodeCursor(u uuid.UUID) string {
	return base64.StdEncoding.EncodeToString(u.Bytes())
}

func DecodeCursor(s string) (uuid.UUID, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromBytes(bytes)
}
