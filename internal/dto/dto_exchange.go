package dto

import "time"

type CreateExchangeRequest struct {
	InitiatorID     uint `json:"initiator_id"`
	RecipientID     uint `json:"recipient_id"`
	InitiatorBookID uint `json:"initiator_book_id"`
	RecipientBookID uint `json:"recipient_book_id"`
}

type ExchangeResponse struct {
	ID              uint       `json:"id"`
	InitiatorID     uint       `json:"initiator_id"`
	RecipientID     uint       `json:"recipient_id"`
	InitiatorBookID uint       `json:"initiator_book_id"`
	RecipientBookID uint       `json:"recipient_book_id"`
	Status          string     `json:"status"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}