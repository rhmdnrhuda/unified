package entity

type MessageRequest struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}
