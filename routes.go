package main

import "fmt"

const (
	apiBasePath       = "/api/v1"
	blobCreatePath    = "/blob"
	blobPath          = "/blob/{id}"
	blobSegmentsPath  = "/blob/{id}/segments"
	profileCreatePath = "/profile"
	profilePath       = "/profile/{id}"
	tagsPath          = "/profile/{id}/tags"
	segmentPath       = "/profile/{id}/segment/{segmentType}"
	categoriesPath    = "/profile/{id}/segment/{segmentType}/categories"
	topCategoriesPath = "/profile/{id}/segment/{segmentType}/topcategories"
)

func (s *server) setupRoutes() {
	s.router.HandleFunc(fmt.Sprintf("PUT %s%s", apiBasePath, profileCreatePath), handleUpsertProfile(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("PUT %s%s", apiBasePath, blobCreatePath), handleUpsertBlob(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, profilePath), handleGetProfile(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, segmentPath), handleGetSegment(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, categoriesPath), handleGetCategories(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, tagsPath), handleGetTags(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, topCategoriesPath), handleGetTopCategories(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, blobPath), handleGetBlob(s.db, s.log))
	s.router.HandleFunc(fmt.Sprintf("GET %s%s", apiBasePath, blobSegmentsPath), handleGetSegmentsFromBlob(s.db, s.log))
}
