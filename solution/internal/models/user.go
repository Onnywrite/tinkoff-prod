package models

type Profile struct {
	Login       string `json:"login"`
	Email       string `json:"email"`
	CountryCode string `json:"countryCode"`
	IsPublic    bool   `json:"isPublic"`
	Phone       string `json:"phone"`
	Image       string `json:"image"`
}

type User struct {
	Login       string `validate:"required,ne=me" json:"login"`
	Email       string `validate:"required,email" json:"email"`
	CountryCode string `validate:"required,len=2" json:"countryCode"`
	IsPublic    bool   `validate:"required" json:"isPublic"`
	Phone       string `validate:"required,e164" json:"phone"`
	Image       string `validate:"required,http_url,max=100" json:"image"`
	Password    string `validate:"required,min=8" json:"password"`
}
