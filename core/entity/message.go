package entity

type MessageRequest struct {
	EventType   string      `json:"eventType"`
	FromNo      string      `json:"fromNo"`
	Platform    string      `json:"platform"`
	AccountNo   string      `json:"accountNo"`
	AccountName string      `json:"accountName"`
	Data        DataRequest `json:"data"`
}

type DataRequest struct {
	ID        string `json:"id"`
	CustNo    string `json:"custNo"`
	CustName  string `json:"custName"`
	Type      string `json:"type"`
	Text      string `json:"text"`
	TimeStamp string `json:"timeStamp"`
}

type AdaRequest struct {
	Platform string `json:"platform"`
	From     string `json:"from"`
	To       string `json:"to"`
	Type     string `json:"type"`
	Text     string `json:"text"`
}

type AdaButtonRequest struct {
	Platform   string   `json:"platform"`
	From       string   `json:"from"`
	To         string   `json:"to"`
	Type       string   `json:"type"`
	Text       string   `json:"text"`
	HeaderType string   `json:"headerType"`
	Header     string   `json:"header"`
	Footer     string   `json:"footer"`
	Buttons    []string `json:"buttons"`
}

type UserTemporaryData struct {
	UniversityPreferences []string `json:"university_preferences"`
	MajorPreferences      []string `json:"major_preferences"`
	Feature               string   `json:"feature"`
}
