package ddb

import (
	"personalisation-poc/repository"
	"time"

	"github.com/guregu/dynamo/v2"
)

var _ repository.ProfilesRepo = &DB{} // compile time check

type Option func(*DB)

// WithTTL updates the items' Time To Live. By default it's zero (unlimited).
func WithTTL(ttl time.Duration) Option {
	return func(db *DB) {
		db.ttl = ttl
	}
}

// DB implements the ProfilesRepo interface backed by a DynamoDB table.
// It follows the principles of Single Table Design.
type DB struct {
	table dynamo.Table
	ttl   time.Duration
}

// NewDB returns a new DynamoDB-backed implementation of the ProfilesRepo interface.
// It uses a single DynamoDB table.
func NewDB(table dynamo.Table, opts ...Option) *DB {
	db := &DB{
		table: table,
		ttl:   0,
	}
	for _, opt := range opts {
		opt(db)
	}

	return db
}
