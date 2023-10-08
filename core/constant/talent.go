package constant

import "github.com/rhmdnrhuda/unified/core/entity"

const (
	Chat           = 1
	Teleconference = 2
	Offline        = 3

	ChatStr           = "Teks"
	TeleconferenceStr = "Call"
	OfflineStr        = "Offline"
)

var (
	ServiceMap = map[string]int{
		ChatStr:           Chat,
		TeleconferenceStr: Teleconference,
		OfflineStr:        Offline,
	}

	ServiceMapStr = map[int]string{
		Chat:           ChatStr,
		Teleconference: TeleconferenceStr,
		Offline:        OfflineStr,
	}

	TalentMap = make(map[int64]entity.TalentRequest)
	OTPMap    = make(map[int64]int64)
)
