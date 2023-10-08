package constant

type State string

const (
	IDLE        State = "idle"
	UNI_CHECK   State = "uni_check"
	UNI_BUDDY   State = "uni_buddy"
	UNI_CONNECT State = "uni_connect"
	UNI_ALERT   State = "uni_alert"
)

var ()
