package ddb

import (
	"personalisation-poc/model"
	"time"

	"github.com/samber/lo"
)

const segmentItemKeyPrefix = "SEG"

type segment struct {
	PK            string     `dynamo:"pk,hash"`  // partition key
	SK            string     `dynamo:"sk,range"` // sort key
	ItemType      string     `dynamo:"typ"`      // item type
	SegmentType   string     `dynamo:"seg_typ"`
	Categories    []category `dynamo:"cats"`
	TopCategories []string   `dynamo:"top_cats,set,omitempty"`
	CreatedAt     time.Time  `dynamo:"created_at"`
	UpdatedAt     time.Time  `dynamo:"updated_at"`
	TTL           int64      `dynamo:"ttl,unixtime"` // TTL for the segment
}

type category struct {
	ID    string  `dynamo:"id"`
	Score float64 `dynamo:"score"`
}

func toCanonicalSegment(dbModel segment) *model.Segment {
	return &model.Segment{
		Type: dbModel.SegmentType,
		Categories: lo.Map(dbModel.Categories, func(c category, _ int) model.Category {
			return model.Category{
				ID:    c.ID,
				Score: c.Score,
			}
		}),
		TopCategories: dbModel.TopCategories,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
		ExpiresAt:     time.Unix(dbModel.TTL, 0),
	}
}

func toDBSegment(seg model.Segment, profileID, segmentType string) segment {
	return segment{
		PK:          buildPK(profileID),
		SK:          buildSK(segmentItemKeyPrefix, segmentType, &seg.CreatedAt),
		ItemType:    segmentItemKeyPrefix,
		SegmentType: seg.Type,
		Categories: lo.Map(seg.Categories, func(c model.Category, _ int) category {
			return category{
				ID:    c.ID,
				Score: c.Score,
			}
		}),
		TopCategories: seg.TopCategories,
		CreatedAt:     seg.CreatedAt,
		UpdatedAt:     seg.UpdatedAt,
		TTL:           seg.ExpiresAt.Unix(),
	}
}
