package models

import "time"

type Like struct {
	User    User      `json:"user"`
	Post    Post      `json:"post"`
	LikedAt time.Time `json:"liked_at"`
}
