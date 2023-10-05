package entity

type Talent struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	CalendarURL string `json:"calendar_url"`
	University  string `json:"university"`
	Major       string `json:"major"`
	Status      string `json:"status"`
	CommonRepository
}

type TalentRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	CalendarURL string `json:"calendar_url"`
	University  string `json:"university"`
	Major       string `json:"major"`
	Status      string `json:"status"`
}

type TalentResponse struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Username     string        `json:"username"`
	Phone        string        `json:"phone"`
	Website      string        `json:"website"`
	Description  string        `json:"description"`
	Visited      int64         `json:"visited"`
	JoinDate     string        `json:"join_date"`
	ImageURL     string        `json:"image_url"`
	Headline     string        `json:"headline"`
	Address      string        `json:"address"`
	YOE          string        `json:"yoe" ` // years of experience
	Expertise    []string      `json:"expertise"`
	Satisfied    int64         `json:"satisfied"`
	Disappointed int64         `json:"disappointed"`
	PricingRules []PricingRule `json:"pricing_rules"`
	Service      []string      `json:"service"`
	IsActive     bool          `json:"is_active"`
	IsVerified   bool          `json:"is_verified"`
	VerifiedBy   string        `json:"verified_by"`
}

type PricingRule struct {
	Service   string `json:"service"`
	ServiceID int    `json:"service_id"`
	PriceData Price  `json:"price_data"`
}

type Price struct {
	OriginalPrice          int64  `json:"price"`
	OriginalPromotionPrice *int64 `json:"promotion_price,omitempty"`
	PriceString            string `json:"price_string"`
	PromotionPriceString   string `json:"promotion_price_string"`
}

type SearchRequest struct {
	Value  string `json:"value,omitempty"`
	Filter Filter `json:"filter,omitempty"`
	Limit  int64  `json:"limit,omitempty"`
	Page   int64  `json:"page,omitempty"`
	SortBy string `json:"sort_by,omitempty"`
}

type Filter struct {
	MinPrice      int64 `json:"min_price,omitempty"`
	MaxPrice      int64 `json:"max_price,omitempty"`
	MinThumbsUp   int64 `json:"min_thumbs_up,omitempty"`
	MaxThumbsUp   int64 `json:"max_thumbs_up,omitempty"`
	Status        bool  `json:"status,omitempty"`
	PromotionOnly bool  `json:"promotion_only"`
}
