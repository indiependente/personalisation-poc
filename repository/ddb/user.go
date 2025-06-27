package ddb

import (
	"personalisation-poc/model"
	"time"
)

const userItemKeyPrefix = "USER"

type user struct {
	PK        string    `dynamo:"pk,hash"`  // partition key
	SK        string    `dynamo:"sk,range"` // sort key
	ItemType  string    `dynamo:"typ"`      // item type
	ID        string    `dynamo:"id"`
	Tags      []string  `dynamo:"tags,set,omitempty"`
	CreatedAt time.Time `dynamo:"created_at"`
	UpdatedAt time.Time `dynamo:"updated_at"`
	TTL       int64     `dynamo:"ttl,unixtime"`
}

func toDBUser(p model.Profile) user {
	id := p.ID.String()

	return user{
		PK:        buildPK(id),
		SK:        buildSK(userItemKeyPrefix, id, nil),
		ItemType:  userItemKeyPrefix,
		ID:        id,
		Tags:      p.Tags,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		TTL:       p.ExpiresAt.Unix(),
	}
}
