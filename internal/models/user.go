package models

import "time"

type User struct {
	Id           uint64  `json:"id"`
	Name         string  `json:"name"`
	Lastname     string  `json:"lastname"`
	Email        string  `json:"email"`
	Country      Country `json:"country"`
	IsPublic     bool    `json:"is_public"`
	Image        string  `json:"image"`
	PasswordHash string  `json:"password"`
	Birthday     time.Time
}
