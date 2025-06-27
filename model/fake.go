package model

import (
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

const (
	MorningSegmentType = "morning"
	EveningSegmentType = "evening"
)

var (
	categories = []string{
		"politics",
		"sports",
		"entertainment",
		"technology",
		"world",
		"local",
	}
	categoriesTags = map[string][]string{
		"politics":      {"politics_nerd"},
		"sports":        {"sports_fan"},
		"entertainment": {"binge_watcher", "netflix_addict"},
		"technology":    {"tech_geek", "apple_fanboy", "android_fanboy"},
		"world":         {"world_news_junkie"},
		"local":         {"local_news_junkie"},
	}
)

func FakeProfile() *Profile {
	id := uuid.New()
	now := time.Now()
	p := &Profile{
		ID:   id,
		Tags: []string{},
		Segments: []Segment{
			getRandomSegment(MorningSegmentType, rand.Intn(len(categoriesTags)), now, now),
			getRandomSegment(EveningSegmentType, rand.Intn(len(categoriesTags)), now, now),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	p.Tags = append(p.Tags, categoriesTags[p.Segments[0].TopCategories[0]]...)
	p.Tags = append(p.Tags, categoriesTags[p.Segments[1].TopCategories[0]]...)

	return p
}

func getRandomSegment(typ string, numCategories int, createdAt, updatedAt time.Time) Segment {
	categories := make([]Category, numCategories)
	for i := range numCategories {
		categories[i] = Category{ID: getRandomCategory(), Score: getRandomScore()}
	}
	topCategories := extractTopCategories(categories)

	return Segment{
		Type:          typ,
		Categories:    categories,
		TopCategories: topCategories,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func extractTopCategories(categories []Category) []string {
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Score > categories[j].Score
	})
	topCategories := make([]string, 0, 3) // take top 3 categories if available
	if len(categories) < 3 {
		return lo.Map(categories, func(category Category, _ int) string {
			return category.ID
		})
	}

	for _, category := range categories[:3] {
		topCategories = append(topCategories, category.ID)
	}
	return topCategories
}

func getRandomCategory() string {
	return categories[rand.Intn(len(categories))]
}

func getRandomScore() float64 {
	return rand.Float64()
}
