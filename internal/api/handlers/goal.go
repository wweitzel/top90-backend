package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
)

type GetGoalResponse struct {
	Goal db.Goal `json:"goal"`
}

type GetGoalsRequest struct {
	Pagination db.Pagination     `json:"pagination"`
	Filter     db.GetGoalsFilter `json:"filter"`
}

type GetGoalsResponse struct {
	Goals []db.Goal `json:"goals"`
	Total int       `json:"total"`
}

type GoalHandler struct {
	dao      dao.Top90DAO
	s3Client s3.S3Client
	s3Bucket string
}

const presignedExpirationTime = 10 * time.Minute

func NewGoalHandler(dao dao.Top90DAO, s3Client s3.S3Client, s3Bucket string) *GoalHandler {
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
	goal.PresignedUrl = h.presignedUrl(goal.S3ObjectKey)
	goal.ThumbnailPresignedUrl = h.presignedUrl(string(goal.ThumbnailS3Key))
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
		goals[i].PresignedUrl = h.presignedUrl(goals[i].S3ObjectKey)
		goals[i].ThumbnailPresignedUrl = h.presignedUrl(string(goals[i].ThumbnailS3Key))
	}
	ok(w, GetGoalsResponse{Goals: goals, Total: count})
}

func (h *GoalHandler) presignedUrl(key string) string {
	url, _ := h.s3Client.PresignedUrl(key, h.s3Bucket, presignedExpirationTime)
	return url
}
