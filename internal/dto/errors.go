package dto

import "errors"

var (
	// Book repository errors
	ErrBookCreateFailed = errors.New("error creating book in db")
	ErrBookGetFailed    = errors.New("error getting book from db")
	ErrBookUpdateFailed = errors.New("error updating book in db")
	ErrBookDeleteFailed = errors.New("error deleting book in db")

	// Exchange repository errors
	ErrExchangeCreateFailed   = errors.New("error create exchange in db")
	ErrExchangeUpdateFailed   = errors.New("error update exchange in db")
	ErrExchangeCancelFailed   = errors.New("error cancel exchange in db")
	ErrExchangeCompleteFailed = errors.New("error complete exchange in db")
	ErrExchangeGetFailed      = errors.New("error get exchange in db")

	// Genre repository errors
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrInvalidInput = errors.New("invalid input")

	// Review repository errors
	ErrReviewCreateFail = errors.New("failed to create review")
	ErrReviewNotFound   = errors.New("review not found")
	ErrReviewDeleteFail = errors.New("failed to delete review")

	// User repository errors
	ErrUserCreateFailed = errors.New("failed to create user")
	ErrUserUpdateFailed = errors.New("failed to update user")
	ErrUserDeleteFailed = errors.New("failed to delete user")
	ErrUserGetFailed    = errors.New("failed to get user")

	// Book Service errors
	ErrBookForbidden    = errors.New("forbidden")
	ErrBookInExchange   = errors.New("book is involved in exchange")
	ErrInvalidBookInput = errors.New("invalid book input")
	ErrAISummaryFailed  = errors.New("failed to generate ai summary")

	// Review Service errors
	ErrExchangeInvalidID   = errors.New("invalid exchange id")
	ErrExchangeNotPending  = errors.New("exchange is not pending")
	ErrExchangeNotAccepted = errors.New("exchange is not accepted")
	ErrInitiatorNotOwner   = errors.New("initiator does not own the book")
	ErrRecipientNotOwner   = errors.New("recipient does not own the book")
	ErrUnavailable         = errors.New("initiator book is unavailable")
	ErrRUnavailable        = errors.New("recipient book is unavailable")

	ErrReviewTextRequired    = errors.New("review text is required")
	ErrReviewTextLength      = errors.New("review text must be between 10 and 150 characters")
	ErrInvalidRating         = errors.New("rating must be between 1 and 5")
	ErrSelfReviewForbidden   = errors.New("cannot leave review to yourself")
	ErrReviewDeleteForbidden = errors.New("you are not allowed to delete this review")

	ErrEmailAlreadyUsed        = errors.New("email already in use")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserProfileFailed       = errors.New("failed to get user profile")
	ErrUserExchangesFailed     = errors.New("failed to get user exchanges")
	ErrUserProfileUpdateFailed = errors.New("failed to update user profile")
	ErrUserListFailed          = errors.New("failed to list users")
	ErrUserProfileStatsFailed  = errors.New("failed to calculate user profile stats")
	ErrUserPasswordHashFailed  = errors.New("failed to hash password")
)
