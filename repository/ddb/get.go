package ddb

import (
	"context"
	"encoding/json"
	"fmt"
	"personalisation-poc/model"
	"personalisation-poc/repository"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/guregu/dynamo/v2"
	"github.com/samber/lo"
)

func (d *DB) GetProfileByID(ctx context.Context, id string) (*model.Profile, error) {
	var (
		user     *user
		segments []segment
		item     map[string]types.AttributeValue
	)

	// Get all items for the profile
	iter := d.table.Get(partitionKey, buildPK(id)).Iter()
	for iter.Next(ctx, &item) {
		itemTyp, ok := item[itemType].(*types.AttributeValueMemberS)
		if !ok {
			return nil, fmt.Errorf("invalid sort key")
		}
		switch { // add all supported item types to this switch statement
		case itemTyp == nil:
			return nil, fmt.Errorf("invalid sort key")
		case strings.HasPrefix(itemTyp.Value, userItemKeyPrefix): // user item
			err := dynamo.UnmarshalItem(item, &user)
			if err != nil {
				return nil, fmt.Errorf("unmarshal user: %w", err)
			}
		case strings.HasPrefix(itemTyp.Value, segmentItemKeyPrefix): // segment item
			var segmt segment
			err := dynamo.UnmarshalItem(item, &segmt)
			if err != nil {
				return nil, fmt.Errorf("unmarshal sub profile: %w", err)
			}
			segments = append(segments, segmt)
		default:
			return nil, fmt.Errorf("get profile: unknown item type: %s", *itemTyp)
		}
	}

	// Check for any errors from the iterator
	err := iter.Err()
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, repository.ErrNoProfileFound
	}

	// Convert the items to a canonical profile
	return toCanonicalProfile(*user, segments), nil
}

func (d *DB) GetSegment(ctx context.Context, profileID string, segmentType string, createdAt time.Time) (*model.Segment, error) {
	var segment segment
	err := d.table.Get(partitionKey, buildPK(profileID)).
		Range(sortKey, dynamo.BeginsWith, buildSK(segmentItemKeyPrefix, segmentType, &createdAt)).
		One(ctx, &segment)

	return toCanonicalSegment(segment), err
}

func (d *DB) GetCategories(ctx context.Context, profileID string, segmentType string) ([]model.Category, error) {
	var categories []category
	err := d.table.Get(partitionKey, buildPK(profileID)).
		Range(sortKey, dynamo.Equal, buildSK(segmentItemKeyPrefix, segmentType, nil)).
		Project("cats").
		One(ctx, &categories)

	return lo.Map(categories, func(cat category, _ int) model.Category {
		return model.Category{
			ID:    cat.ID,
			Score: cat.Score,
		}
	}), err
}

func (d *DB) GetUserTags(ctx context.Context, profileID string) ([]string, error) {
	var tags map[string]any
	err := d.table.Get(partitionKey, buildPK(profileID)).
		Range(sortKey, dynamo.Equal, buildSK(userItemKeyPrefix, profileID, nil)).
		Project("tags").
		One(ctx, &tags)

	tagList, ok := tags["tags"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid tags")
	}

	return tagList, err
}

func (d *DB) GetTopCategories(ctx context.Context, profileID string, segmentType string) ([]string, error) {
	var topCategories map[string]any
	err := d.table.Get(partitionKey, buildPK(profileID)).
		Range(sortKey, dynamo.Equal, buildSK(segmentItemKeyPrefix, segmentType, nil)).
		Project("top_cats").
		One(ctx, &topCategories)

	topCategoryList, ok := topCategories["top_cats"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid top categories")
	}

	return topCategoryList, err
}

func (d *DB) GetBlob(ctx context.Context, profileID string) ([]byte, error) {
	var blob blob
	err := d.table.Get(partitionKey, buildPK(profileID)).
		Range(sortKey, dynamo.Equal, buildSK(blobItemKeyPrefix, profileID, nil)).
		Project("rawdata").
		One(ctx, &blob)

	if err != nil {
		return nil, err
	}

	// Marshal the interface{} back to JSON bytes
	return json.Marshal(blob.Data)
}

func (d *DB) GetRawSegmentsFromBlob(ctx context.Context, profileID string) ([]byte, error) {
	// Project only the segments field from rawdata
	var result struct {
		RawData map[string]any `dynamo:"rawdata"`
	}

	err := d.table.Get(partitionKey, buildPK(profileID)).
		Range(sortKey, dynamo.Equal, buildSK(blobItemKeyPrefix, profileID, nil)).
		Project("rawdata.'segments'").
		One(ctx, &result)

	if err != nil {
		return nil, fmt.Errorf("failed to get segments from blob: %w", err)
	}

	// Extract segments from the projected result
	segmentsData, exists := result.RawData["segments"]
	if !exists {
		return nil, repository.ErrNoSegmentsFound // No segments found
	}

	return json.Marshal(segmentsData)
}
