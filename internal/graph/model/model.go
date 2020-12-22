package model

import (
	"encoding/base64"

	"github.com/Tarick/naca-items/internal/entity"
	"github.com/gofrs/uuid"
)

type ItemsConnection struct {
	Items []*entity.Item
	// Slicing it from-to for pagination
	FromIndex int
	ToIndex   int
}
type ItemsEdge struct {
	Node   *entity.Item `json:"node"`
	Cursor string       `json:"cursor"`
}

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	EndCursor       *string `json:"endCursor"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
}

func (i *ItemsConnection) TotalCount() int {
	return len(i.Items)
}

// PageInfo returns PageInfo for paging with encoded cursors
func (i *ItemsConnection) PageInfo() PageInfo {
	var start, end string
	if len(i.Items) > 0 {
		start = EncodeCursor(i.Items[i.FromIndex].UUID)
		end = EncodeCursor(i.Items[i.ToIndex].UUID)
	}
	return PageInfo{
		StartCursor:     &start,
		EndCursor:       &end,
		HasNextPage:     i.ToIndex < len(i.Items)-1,
		HasPreviousPage: i.FromIndex > 0,
	}
}

// EncodeCursor creates base64 representation of UUID bytes array.
// Manual decoding will not be readable, need to convert byte array to string
func EncodeCursor(u uuid.UUID) string {
	return base64.StdEncoding.EncodeToString(u.Bytes())
}

// DecodeCursor decodes cursor
func DecodeCursor(s string) (uuid.UUID, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromBytes(bytes)
}
