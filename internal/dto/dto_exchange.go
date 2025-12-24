package dto

type CreateExchangeRequest struct {
	InitiatorID     uint `json:"initiator_id"`
	RecipientID     uint `json:"recipient_id"`
	InitiatorBookID uint `json:"initiator_book_id"`
	RecipientBookID uint `json:"recipient_book_id"`
}
