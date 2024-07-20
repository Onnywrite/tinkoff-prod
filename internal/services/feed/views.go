package feed

import (
	"regexp"
	"strings"

	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type Author struct {
	Id       uint64 `json:"id"`
	Name     string `json:"name"`
	Lastname string `json:"surname"`
}
type FullPost struct {
	Id          uint64  `json:"id"`
	Author      Author  `json:"author"`
	Content     string  `json:"content"`
	ImageUrl    *string `json:"image_url"`
	PublishedAt string  `json:"published_at"`
	UpdatedAt   *string `json:"updated_at"`
}
type AuthorlessPost struct {
	Id          uint64  `json:"id"`
	Content     string  `json:"content"`
	ImageUrl    *string `json:"image_url"`
	PublishedAt string  `json:"published_at"`
	UpdatedAt   *string `json:"updated_at"`
}
type Page[T any] struct {
	First   uint64 `json:"first"`
	Current uint64 `json:"current"`
	Last    uint64 `json:"last"`
	Posts   []T    `json:"posts"`
}

type PagedFeed Page[FullPost]
type PagedProfileFeed Page[AuthorlessPost]

type Post struct {
	AuthorId   uint64   `json:"author_id"`
	Content    *string  `json:"content"`
	ImagesUrls []string `json:"images_urls"`
}

func (p Post) Validate() ero.Error {
	type fieldFault struct {
		Field   string
		Message string
	}

	spaceRegex := regexp.MustCompile(`^\s*$`)
	urlRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b\/([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)

	faults := make([]fieldFault, 0, 2)

	if p.Content == nil || spaceRegex.MatchString(*p.Content) {
		faults = append(faults, fieldFault{
			Field:   "content",
			Message: "cannot be empty",
		})
	} else {
		*p.Content = strings.TrimSpace(*p.Content)
	}

	if p.ImagesUrls != nil {
		if len(p.ImagesUrls) == 0 {
			p.ImagesUrls = nil
		}
		for i := range p.ImagesUrls {
			p.ImagesUrls[i] = strings.Trim(p.ImagesUrls[i], " ")
			if !urlRegex.MatchString(p.ImagesUrls[i]) {
				faults = append(faults, fieldFault{
					Field:   "images_urls",
					Message: "not all URLs are valid",
				})
				break
			}
		}
	}

	if len(faults) > 0 {
		fields := make([]string, len(faults))
		for i := range faults {
			fields[i] = faults[i].Field
		}
		return ero.NewValidation(erolog.NewContextBuilder().With("fields", fields).Build(), faults)
	}
	return nil
}
