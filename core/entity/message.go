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
	Platform     string   `json:"platform"`
	From         string   `json:"from"`
	To           string   `json:"to"`
	Type         string   `json:"type"`
	Text         string   `json:"text"`
	TemplateName string   `json:"templateName,omitempty"`
	TemplateLang string   `json:"templateLang,omitempty"`
	TemplateData []string `json:"templateData,omitempty"`
	StickerID    string   `json:"stickerId"`
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

type MessageResponse struct {
	//{"status":200,"errorCode":0,"message":"Success","data":["c607f10b-de9c-4be0-8409-71ffd1ecea3d"]}
	Status int64    `json:"status"`
	Data   []string `json:"data"`
}

type Event struct {
	EventTitle string `json:"event_title"`
	Date       string `json:"date"`
}

type AlertResponse struct {
	Events []Event `json:"events"`
}
