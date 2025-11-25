// Package service implements business logic.
package service

import "errors"

var (
	// ErrTeamAlreadyExists indicates that the team already exists.
	ErrTeamAlreadyExists = errors.New("team already exists")
	// ErrTeamNotFound indicates that the team was not found.
	ErrTeamNotFound = errors.New("team not found")

	// ErrUserNotFound indicates that the user was not found.
	ErrUserNotFound = errors.New("user not found")

	// ErrPRNotFound indicates that the pull request was not found.
	ErrPRNotFound = errors.New("pr not found")
	// ErrPRAlreadyExists indicates that the pull request already exists.
	ErrPRAlreadyExists = errors.New("pr already exists")
	// ErrPRMerged indicates that the pull request is already merged.
	ErrPRMerged = errors.New("pr already merged")
	// ErrReviewerNotAssigned indicates that the user is not assigned as a reviewer.
	ErrReviewerNotAssigned = errors.New("reviewer not assigned")
	// ErrNoCandidate indicates that no suitable candidate was found for reassignment.
	ErrNoCandidate = errors.New("no candidate for reassignment")
)
