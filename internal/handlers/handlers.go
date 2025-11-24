package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"ilyaytrewq/PR_assigning_service/internal/api"
	"ilyaytrewq/PR_assigning_service/internal/service"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("PostPullRequestCreate decode:", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	pr, err := h.services.PRs.CreatePR(r.Context(), &body)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound),
			errors.Is(err, service.ErrTeamNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "author or team not found")
		case errors.Is(err, service.ErrPRAlreadyExists):
			h.writeError(w, http.StatusBadRequest, api.PREXISTS, "pull_request_id already exists")
		default:
			log.Println("PostPullRequestCreate:", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(struct {
		Pr *api.PullRequest `json:"pr"`
	}{Pr: pr}); err != nil {
		log.Println("PostPullRequestCreate encode:", err)
	}
}

func (h *Handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("PostPullRequestMerge decode:", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	pr, err := h.services.PRs.MergePR(r.Context(), body.PullRequestId)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPRNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "pull request not found")
		default:
			log.Println("PostPullRequestMerge:", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Pr *api.PullRequest `json:"pr"`
	}{Pr: pr}); err != nil {
		log.Println("PostPullRequestMerge encode:", err)
	}
}

func (h *Handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("PostPullRequestReassign decode:", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	pr, replacedBy, err := h.services.PRs.ReassignReviewer(r.Context(), &body)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPRNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "pull request not found")
		case errors.Is(err, service.ErrPRMerged):
			h.writeError(w, http.StatusConflict, api.PRMERGED, "pull request is already merged")
		case errors.Is(err, service.ErrReviewerNotAssigned):
			h.writeError(w, http.StatusConflict, api.NOTASSIGNED, "user is not assigned as reviewer")
		case errors.Is(err, service.ErrNoCandidate):
			h.writeError(w, http.StatusConflict, api.NOCANDIDATE, "no candidate for reassignment")
		default:
			log.Println("PostPullRequestReassign:", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Pr         *api.PullRequest `json:"pr"`
		ReplacedBy string           `json:"replaced_by"`
	}{
		Pr:         pr,
		ReplacedBy: replacedBy,
	}); err != nil {
		log.Println("PostPullRequestReassign encode:", err)
	}
}

func (h *Handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var team api.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		log.Println("PostTeamAdd decode:", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	if team.TeamName == "" {
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "team name is required")
		return
	}

	if err := h.services.Teams.AddTeam(r.Context(), &team); err != nil {
		switch {
		case errors.Is(err, service.ErrTeamAlreadyExists):
			h.writeError(w, http.StatusBadRequest, api.TEAMEXISTS, "team_name already exists")
		default:
			log.Println("PostTeamAdd:", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(struct {
		Team *api.Team `json:"team"`
	}{Team: &team}); err != nil {
		log.Println("PostTeamAdd encode:", err)
	}
}

func (h *Handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	teamName := string(params.TeamName)

	team, err := h.services.Teams.GetTeam(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTeamNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "team not found")
		default:
			log.Println("GetTeamGet:", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Team *api.Team `json:"team"`
	}{Team: team}); err != nil {
		log.Println("GetTeamGet encode:", err)
	}
}

func (h *Handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	userID := string(params.UserId)

	prs, err := h.services.Users.GetReviewPullRequests(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "user not found")
			return
		}
		log.Println("GetUsersGetReview:", err)
		h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		UserID       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       userID,
		PullRequests: toShorts(prs),
	}); err != nil {
		log.Println("GetUsersGetReview encode:", err)
	}
}

func (h *Handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body api.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("PostUsersSetIsActive decode:", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	user, err := h.services.Users.SetIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "user not found")
		default:
			log.Println("PostUsersSetIsActive:", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		User *api.User `json:"user"`
	}{User: user}); err != nil {
		log.Println("PostUsersSetIsActive encode:", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code api.ErrorResponseErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := api.ErrorResponse{
		Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("writeError encode:", err)
	}
}

func toShorts(prs []*api.PullRequest) []api.PullRequestShort {
	prshs := []api.PullRequestShort{}
	for _, pr := range prs {
		prshs = append(prshs, toShort(pr))
	}
	return prshs
}

func toShort(pr *api.PullRequest) api.PullRequestShort {
	return api.PullRequestShort{
		PullRequestId:   pr.PullRequestId,
		PullRequestName: pr.PullRequestName,
		AuthorId:        pr.AuthorId,
		Status:          api.PullRequestShortStatus(pr.Status),
	}
}
