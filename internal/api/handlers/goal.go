package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
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

func NewGoalHandler(dao db.Top90DAO, s3Client s3.S3Client, s3Bucket string) *GoalHandler {
	return &GoalHandler{
		dao:      dao,
		s3Client: s3Client,
		s3Bucket: s3Bucket,
	}
}

func (h *GoalHandler) GetGoal(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	goal, err := h.dao.GetGoal(id)
	if err != nil {
		log.Println(err)
	}

	goal.PresignedUrl = h.s3Client.PresignedUrl(goal.S3ObjectKey, h.s3Bucket)
	goal.ThumbnailPresignedUrl = h.s3Client.PresignedUrl(goal.ThumbnailS3Key, h.s3Bucket)

	respond(w, http.StatusOK, GetGoalResponse{
		Goal: goal,
	})
}

func (h *GoalHandler) GetGoals(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetGoalsRequest](json)
	if err != nil {
		respond(w, http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	if request.Pagination.Limit == 0 {
		request.Pagination.Limit = 10
	}

	count, err := h.dao.CountGoals(request.Filter)
	if err != nil {
		log.Println(err)
	}

	goals, err := h.dao.GetGoals(request.Pagination, request.Filter)
	if err != nil {
		log.Println(err)
	}

	for i := range goals {
		goals[i].PresignedUrl = h.s3Client.PresignedUrl(goals[i].S3ObjectKey, h.s3Bucket)
		goals[i].ThumbnailPresignedUrl = h.s3Client.PresignedUrl(goals[i].ThumbnailS3Key, h.s3Bucket)
	}

	respond(w, http.StatusOK, GetGoalsResponse{
		Goals: goals,
		Total: count,
	})
}
