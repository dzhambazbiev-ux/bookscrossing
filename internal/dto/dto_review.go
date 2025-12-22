package dto

type CreateReviewRequest struct {
	TargetUserID uint   `json:"target_user_id"`
	TargetBookID uint   `json:"target_book_id"`
	Text         string `json:"text"`
	Rating       int    `json:"rating"`
}
