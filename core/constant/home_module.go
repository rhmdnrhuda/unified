package constant

const (
	// Home Module Type
	TypeBanner       = 1
	TypeContent      = 2
	TypeCategory     = 3
	TypeQuery        = 4
	TypeOrderHistory = 5

	TypeBannerStr       = "BANNER"
	TypeContentStr      = "CONTENT"
	TypeCategoryStr     = "CATEGORY"
	TypeQueryStr        = "QUERY"
	TypeOrderHistoryStr = "ORDER_HISTORY"
)

var (
	MapType = map[string]int64{
		TypeBannerStr:       TypeBanner,
		TypeContentStr:      TypeContent,
		TypeCategoryStr:     TypeCategory,
		TypeQueryStr:        TypeQuery,
		TypeOrderHistoryStr: TypeOrderHistory,
	}

	MapTypeString = map[int64]string{
		TypeBanner:       TypeBannerStr,
		TypeContent:      TypeContentStr,
		TypeCategory:     TypeCategoryStr,
		TypeQuery:        TypeQueryStr,
		TypeOrderHistory: TypeOrderHistoryStr,
	}
)
