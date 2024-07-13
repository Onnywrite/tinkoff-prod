package models

import "time"

type Post struct {
	Id          uint64     `json:"id"`
	Author      User       `json:"author"`
	Content     string     `json:"content"`
	ImagesUrls  *[]string  `json:"images_urls"`
	PublishedAt time.Time  `json:"published_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
