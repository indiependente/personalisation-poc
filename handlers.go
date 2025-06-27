package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"personalisation-poc/model"
	"personalisation-poc/repository"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

const (
	idQueryParam        = "id"
	segmentQueryParam   = "segmentType"
	createdAtQueryParam = "createdAt"
)

func handleUpsertProfile(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("upserting profile")
		data, err := io.ReadAll(r.Body)
		if err != nil {
			httpError(w, log, err, "error reading body", http.StatusBadRequest)
			return
		}
		log.Debug("body", "body", string(data))

		var profile model.Profile
		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&profile); err != nil {
			httpError(w, log, err, "error decoding profile", http.StatusBadRequest)
			return
		}
		log.Debug("profile decoded", "profile", profile)
		validateUpsertProfile(&profile)

		if err := repo.UpsertProfile(r.Context(), profile); err != nil {
			httpError(w, log, err, "error upserting profile", http.StatusInternalServerError)
			return
		}
		log.Debug("profile upserted", "profile", profile)
		w.WriteHeader(http.StatusCreated)
	}
}

func handleUpsertBlob(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("upserting blob")
		data, err := io.ReadAll(r.Body)
		if err != nil {
			httpError(w, log, err, "error reading body", http.StatusBadRequest)
			return
		}
		var profile model.Profile
		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&profile); err != nil {
			httpError(w, log, err, "error decoding profile", http.StatusBadRequest)
			return
		}
		if err := repo.UpsertBlob(r.Context(), profile.ID.String(), data); err != nil {
			httpError(w, log, err, "error upserting blob", http.StatusInternalServerError)
			return
		}

		log.Debug("blob upserted", "blob", data)
		w.WriteHeader(http.StatusCreated)
	}
}

func validateUpsertProfile(profile *model.Profile) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}
	if profile.UpdatedAt.IsZero() {
		profile.UpdatedAt = time.Now()
	}
	if profile.ExpiresAt.IsZero() {
		profile.ExpiresAt = time.Now().AddDate(1, 0, 0) // 1 year
	}
	profile.Segments = lo.Map(profile.Segments, func(segment model.Segment, _ int) model.Segment {
		if segment.CreatedAt.IsZero() {
			segment.CreatedAt = time.Now()
		}
		if segment.UpdatedAt.IsZero() {
			segment.UpdatedAt = time.Now()
		}
		if segment.ExpiresAt.IsZero() {
			segment.ExpiresAt = time.Now().AddDate(0, 6, 0) // 6 months
		}
		return segment
	})
}

func handleGetProfile(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		profile, err := repo.GetProfileByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNoProfileFound) {
				httpError(w, log, err, "profile not found", http.StatusNotFound)
				return
			}
			log.Error("error getting profile", "error", err)
			httpError(w, log, err, "error getting profile", http.StatusInternalServerError)
			return
		}
		log.Debug("profile retrieved", "profile", profile)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}
}

func handleGetSegment(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		segmentType := r.PathValue(segmentQueryParam)
		if segmentType == "" {
			httpError(w, log, errors.New("segmentType is required"), "segmentType is required", http.StatusBadRequest)
			return
		}
		var createdAt time.Time
		timestamp := r.URL.Query().Get(createdAtQueryParam)
		if timestamp != "" {
			t, err := time.Parse(time.RFC3339, timestamp)
			if err != nil {
				httpError(w, log, err, "failed parsing created at timestamp", http.StatusBadRequest)
				return
			}
			createdAt = t
		}

		segment, err := repo.GetSegment(r.Context(), id, segmentType, createdAt)
		if err != nil {
			if errors.Is(err, repository.ErrNoSegmentsFound) {
				httpError(w, log, err, "segment not found", http.StatusNotFound)
				return
			}
			httpError(w, log, err, "error getting segment", http.StatusInternalServerError)
			return
		}
		log.Debug("segment retrieved", "segment", segment)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(segment)
	}
}

func handleGetCategories(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		segmentType := r.PathValue(segmentQueryParam)
		if segmentType == "" {
			httpError(w, log, errors.New("segmentType is required"), "segmentType is required", http.StatusBadRequest)
			return
		}

		categories, err := repo.GetCategories(r.Context(), id, segmentType)
		if err != nil {
			httpError(w, log, err, "error getting categories", http.StatusInternalServerError)
			return
		}
		log.Debug("categories retrieved", "categories", categories)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	}
}

func handleGetTopCategories(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		segmentType := r.PathValue(segmentQueryParam)
		if segmentType == "" {
			httpError(w, log, errors.New("segmentType is required"), "segmentType is required", http.StatusBadRequest)
			return
		}

		topCategories, err := repo.GetTopCategories(r.Context(), id, segmentType)
		if err != nil {
			httpError(w, log, err, "error getting top categories", http.StatusInternalServerError)
			return
		}
		log.Debug("top categories retrieved", "topCategories", topCategories)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topCategories)
	}
}

func handleGetTags(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		tags, err := repo.GetUserTags(r.Context(), id)
		if err != nil {
			httpError(w, log, err, "error getting tags", http.StatusInternalServerError)
			return
		}
		log.Debug("tags retrieved", "tags", tags)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tags)
	}
}

func handleGetSegmentsFromBlob(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		segments, err := repo.GetRawSegmentsFromBlob(r.Context(), id)
		if err != nil {
			httpError(w, log, err, "error getting segments from blob", http.StatusInternalServerError)
			return
		}
		log.Debug("segments retrieved", "segments", segments)
		w.Header().Set("Content-Type", "application/json")
		w.Write(segments)
	}
}

func httpError(w http.ResponseWriter, log *slog.Logger, err error, errMsg string, status int) {
	log.Error(errMsg, "error", err)
	http.Error(w, err.Error(), status)
}

func handleGetBlob(repo repository.ProfilesRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpError(w, log, errors.New("id is required"), "id is required", http.StatusBadRequest)
			return
		}

		blob, err := repo.GetBlob(r.Context(), id)
		if err != nil {
			httpError(w, log, err, "error getting blob", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(blob)
	}
}
