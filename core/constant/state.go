package constant

type State string
type RegisterState string

const (
	IDLE      State = "idle"
	Start     State = "start"
	Register  State = "register"
	Order     State = "order"
	Balance   State = "balance"
	Update    State = "update"
	ResendOTP State = "resend_otp"

	Cancel = "cancel"

	None              RegisterState = ""
	Name              RegisterState = "name"
	Username          RegisterState = "username"
	Phone             RegisterState = "phone"
	Email             RegisterState = "email"
	Website           RegisterState = "website"
	Image             RegisterState = "image"
	CareerStartDate   RegisterState = "career_start_date"
	Expertise         RegisterState = "expertise"
	Headline          RegisterState = "headline"
	Description       RegisterState = "description"
	Pricing           RegisterState = "pricing"
	EmailVerification RegisterState = "email_verification"
	Registered        RegisterState = "registered"
	Status            RegisterState = "status"
)

var (
	CommandState      = make(map[int64]State)         // key: chatID
	RegistrationState = make(map[int64]RegisterState) // key: user telegram ID
	UpdateState       = make(map[int64]RegisterState) // key: user telegram ID
)
