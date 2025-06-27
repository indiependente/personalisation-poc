package repository

import (
	"context"
	"errors"
	"time"

	"personalisation-poc/model"
)

var (
	ErrNoSegmentsFound = errors.New("no segments found")
	ErrNoProfileFound  = errors.New("no profile found")
)

type ProfilesRepo interface {
	GetterProfileRepo
	UpserterProfileRepo
}

type GetterProfileRepo interface {
	GetProfileByID(ctx context.Context, id string) (*model.Profile, error)
	GetSegment(ctx context.Context, profileID string, segmentType string, createdAt time.Time) (*model.Segment, error)
	GetCategories(ctx context.Context, profileID string, segmentType string) ([]model.Category, error)
	GetUserTags(ctx context.Context, profileID string) ([]string, error)
	GetTopCategories(ctx context.Context, profileID string, segmentType string) ([]string, error)
	GetBlob(ctx context.Context, profileID string) ([]byte, error)
	GetRawSegmentsFromBlob(ctx context.Context, profileID string) ([]byte, error)
}

type UpserterProfileRepo interface {
	UpsertProfile(ctx context.Context, profile model.Profile) error
	UpsertBlob(ctx context.Context, profileID string, data []byte) error
}
