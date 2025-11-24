package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

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
	start := time.Now()
	var body api.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("PostPullRequestCreate decode error: %v", err)
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
			log.Printf("PostPullRequestCreate internal error: %v", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(struct {
		Pr *api.PullRequest `json:"pr"`
	}{Pr: pr}); err != nil {
		log.Printf("PostPullRequestCreate encode error: %v", err)
	}
	log.Printf("PostPullRequestCreate success: author_id=%s pr_id=%s duration=%s", body.AuthorId, body.PullRequestId, time.Since(start))
}

func (h *Handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var body api.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("PostPullRequestMerge decode error: %v", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	pr, err := h.services.PRs.MergePR(r.Context(), body.PullRequestId)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPRNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "pull request not found")
		default:
			log.Printf("PostPullRequestMerge internal error: %v", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Pr *api.PullRequest `json:"pr"`
	}{Pr: pr}); err != nil {
		log.Printf("PostPullRequestMerge encode error: %v", err)
	}
	log.Printf("PostPullRequestMerge success: pr_id=%s duration=%s", body.PullRequestId, time.Since(start))
}

func (h *Handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var body api.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("PostPullRequestReassign decode error: %v", err)
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
			log.Printf("PostPullRequestReassign internal error: %v", err)
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
		log.Printf("PostPullRequestReassign encode error: %v", err)
	}
	log.Printf("PostPullRequestReassign success: pr_id=%s old_user=%s new_user=%s duration=%s", body.PullRequestId, body.OldUserId, replacedBy, time.Since(start))
}

func (h *Handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var team api.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		log.Printf("PostTeamAdd decode error: %v", err)
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
			log.Printf("PostTeamAdd error: %v", err)
			h.writeError(w, http.StatusBadRequest, api.TEAMEXISTS, "team_name already exists")
		default:
			log.Printf("PostTeamAdd internal error: %v", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(struct {
		Team *api.Team `json:"team"`
	}{Team: &team}); err != nil {
		log.Printf("PostTeamAdd encode error: %v", err)
	}
	log.Printf("PostTeamAdd success: team_name=%s members=%d duration=%s", team.TeamName, len(team.Members), time.Since(start))
}

func (h *Handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	start := time.Now()
	teamName := string(params.TeamName)

	team, err := h.services.Teams.GetTeam(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTeamNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "team not found")
		default:
			log.Printf("GetTeamGet internal error: %v", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Team *api.Team `json:"team"`
	}{Team: team}); err != nil {
		log.Printf("GetTeamGet encode error: %v", err)
	}
	log.Printf("GetTeamGet success: team_name=%s members=%d duration=%s", teamName, len(team.Members), time.Since(start))
}

func (h *Handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	start := time.Now()
	userID := string(params.UserId)

	prs, err := h.services.Users.GetReviewPullRequests(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "user not found")
			return
		}
		log.Printf("GetUsersGetReview internal error: %v", err)
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
		log.Printf("GetUsersGetReview encode error: %v", err)
	}
	log.Printf("GetUsersGetReview success: user_id=%s prs=%d duration=%s", userID, len(prs), time.Since(start))
}

func (h *Handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var body api.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("PostUsersSetIsActive decode error: %v", err)
		h.writeError(w, http.StatusBadRequest, api.NOTFOUND, "invalid request body")
		return
	}

	user, err := h.services.Users.SetIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			h.writeError(w, http.StatusNotFound, api.NOTFOUND, "user not found")
		default:
			log.Printf("PostUsersSetIsActive internal error: %v", err)
			h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		User *api.User `json:"user"`
	}{User: user}); err != nil {
		log.Printf("PostUsersSetIsActive encode error: %v", err)
	}
	log.Printf("PostUsersSetIsActive success: user_id=%s is_active=%t duration=%s", body.UserId, body.IsActive, time.Since(start))
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	stats, err := h.services.GetStats(r.Context())
	if err != nil {
		log.Printf("GetStats internal error: %v", err)
		h.writeError(w, http.StatusInternalServerError, api.NOTFOUND, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("GetStats encode error: %v", err)
	}
	log.Printf("GetStats success: duration=%s", time.Since(start))
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
		log.Printf("writeError encode error: %v", err)
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
