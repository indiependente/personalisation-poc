package main

import (
	"log/slog"
	"net/http"
	"personalisation-poc/repository"
)

type server struct {
	router *http.ServeMux
	db     repository.ProfilesRepo
	log    *slog.Logger
}

func newServer(db repository.ProfilesRepo, log *slog.Logger) *server {
	s := &server{
		router: http.NewServeMux(),
		db:     db,
		log:    log,
	}
	s.setupRoutes()

	return s
}
