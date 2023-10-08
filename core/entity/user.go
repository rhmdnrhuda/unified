package entity

type User struct {
	ID                    int64    `json:"id"`
	Name                  string   `json:"name"`
	Number                string   `json:"number"`
	UniversityPreferences []string `json:"university_preferences"`
	MajorPreferences      []string `json:"major_preferences"`
}
