package entity

type Alert struct {
	ID      int64  `json:"id"`
	UserID  string `json:"user_id"`
	Date    int64  `json:"date"`
	Message string `json:"message"`
}
