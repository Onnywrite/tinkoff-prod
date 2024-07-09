package models

type Profile struct {
	Login       string `json:"login"`
	Email       string `json:"email"`
	CountryCode string `json:"countryCode"`
	IsPublic    bool   `json:"is_public"`
	Phone       string `json:"phone"`
	Image       string `json:"image"`
}

type User struct {
	Login       string `validate:"required,ne=me" json:"login"`
	Email       string `validate:"required,email" json:"email"`
	CountryCode string `validate:"required,len=2" json:"countryCode"`
	IsPublic    bool   `validate:"required" json:"is_public"`
	Phone       string `validate:"required,e164" json:"phone"`
	Image       string `validate:"required,http_url,max=100" json:"image"`
	Password    string `validate:"required,min=8" json:"password"`
}

func (u *User) Profile() *Profile {
	return &Profile{
		Login:       u.Login,
		Email:       u.Email,
		CountryCode: u.CountryCode,
		IsPublic:    u.IsPublic,
		Phone:       u.Phone,
		Image:       u.Image,
	}
}
