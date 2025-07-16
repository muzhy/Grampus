package booking

type CommodityType string

// 商品的类型，目前仅支持货币，不允许用户自定义
const (
	Currency CommodityType = "currency" // 货币
)

type CurrencyType string

const (
	CNY CurrencyType = "CNY"
	USD CurrencyType = "USD"
	EUR CurrencyType = "EUR"
	JPY CurrencyType = "JPY"
)

type Commodity struct {
	ID          string `json:"id" db:"id"`
	Type        string `json:"type" db:"type"`
	Mnemonic    string `json:"mnemonic" db:"mnemonic"`
	Fullname    string `json:"fullname" db:"fullname"`
	Fraction    int32  `json:"fraction" db:"fraction"`
	QuoteFlag   int32  `json:"quote_flag" db:"quote_flag"`
	QuoteSource string `json:"quote_source" db:"quote_source"`
}

type Price struct {
	ID         string `json:"id" db:"id"`
	CommID     string `json:"comm_id" db:"comm_id"`
	CurrencyID string `json:"currency_id" db:"currency_id"`
	Date       string `json:"date" db:"date"`
	Source     string `json:"source" db:"source"`
	Type       string `json:"type" db:"type"`
	ValueNum   int64  `json:"value_num" db:"value_num"`
	ValueDenom int64  `json:"value_denom" db:"value_denom"`
}
