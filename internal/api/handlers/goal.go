package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
)

type GetGoalResponse struct {
	Goal top90.Goal `json:"goal"`
}

type GetGoalsRequest struct {
	Pagination db.Pagination     `json:"pagination"`
	Filter     db.GetGoalsFilter `json:"filter"`
}

type GetGoalsResponse struct {
	Goals []top90.Goal `json:"goals"`
	Total int          `json:"total"`
}

type GoalHandler struct {
	dao      db.Top90DAO
	s3Client s3.S3Client
	s3Bucket string
}

const presignedExpirationTime = 10 * time.Minute

func NewGoalHandler(dao db.Top90DAO, s3Client s3.S3Client, s3Bucket string) *GoalHandler {
	return &GoalHandler{
		dao:      dao,
		s3Client: s3Client,
		s3Bucket: s3Bucket,
	}
}

func (h *GoalHandler) GetGoal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	goal, err := h.dao.GetGoal(id)
	if err != nil {
		internalServerError(w, err)
		return
	}

	goal.PresignedUrl, _ = h.s3Client.PresignedUrl(goal.S3ObjectKey, h.s3Bucket, presignedExpirationTime)
	goal.ThumbnailPresignedUrl, _ = h.s3Client.PresignedUrl(goal.ThumbnailS3Key, h.s3Bucket, presignedExpirationTime)

	ok(w, GetGoalResponse{Goal: goal})
}

func (h *GoalHandler) GetGoals(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetGoalsRequest](json)
	if err != nil {
		internalServerError(w, err)
		return
	}

	if request.Pagination.Limit == 0 {
		request.Pagination.Limit = 10
	}

	count, err := h.dao.CountGoals(request.Filter)
	if err != nil {
		internalServerError(w, err)
		return
	}

	goals, err := h.dao.GetGoals(request.Pagination, request.Filter)
	if err != nil {
		internalServerError(w, err)
		return
	}

	for i := range goals {
		goals[i].PresignedUrl, _ = h.s3Client.PresignedUrl(goals[i].S3ObjectKey, h.s3Bucket, presignedExpirationTime)
		goals[i].ThumbnailPresignedUrl, _ = h.s3Client.PresignedUrl(goals[i].ThumbnailS3Key, h.s3Bucket, presignedExpirationTime)
	}

	ok(w, GetGoalsResponse{Goals: goals, Total: count})
}
