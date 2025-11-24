package service

import "errors"

var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")

	ErrUserNotFound = errors.New("user not found")

	ErrPRNotFound          = errors.New("pr not found")
	ErrPRAlreadyExists     = errors.New("pr already exists")
	ErrPRMerged            = errors.New("pr already merged")
	ErrReviewerNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate         = errors.New("no candidate for reassignment")
)
