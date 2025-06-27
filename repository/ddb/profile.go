package ddb

import (
	"personalisation-poc/model"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

func toCanonicalProfile(u user, ss []segment) *model.Profile {
	return &model.Profile{
		ID:   uuid.MustParse(u.ID),
		Tags: u.Tags,
		Segments: lo.Map(ss, func(s segment, _ int) model.Segment {
			return *toCanonicalSegment(s)
		}),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		ExpiresAt: time.Unix(u.TTL, 0),
	}
}

func toDBItems(p model.Profile) (user, []segment) {
	return toDBUser(p), lo.Map(p.Segments, func(s model.Segment, _ int) segment {
		return toDBSegment(s, p.ID.String(), s.Type)
	})
}
