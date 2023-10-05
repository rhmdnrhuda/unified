package entity

type CommonRepository struct {
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	UpdatedBy string `json:"updated_by"`
	IsDeleted int64  `json:"is_deleted"`
}

type MandatoryRequest struct {
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Version int64  `json:"version"`
}
