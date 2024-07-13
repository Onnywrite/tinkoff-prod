package models

import (
	"fmt"
	"strings"
	"time"
)

type Post struct {
	Id          uint64      `json:"id"`
	Author      User        `json:"author"`
	Content     string      `json:"content"`
	ImagesUrls  StringSlice `json:"images_urls"`
	PublishedAt time.Time   `json:"published_at"`
	UpdatedAt   *time.Time  `json:"updated_at"`
}

type StringSlice []string

func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		s = nil
		return nil
	}
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to scan images_urls")
	}
	str = strings.Trim(str, "{} ")
	*s = strings.Split(str, ",")
	return nil
}
