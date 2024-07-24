package feed

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Onnywrite/tinkoff-prod/internal/services/likes"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type Author struct {
	Id       uint64 `json:"id"`
	Name     string `json:"name"`
	Lastname string `json:"surname"`
	Image    string `json:"image"`
}

// TODO: getting all feed with and without likes if it makes sence
type Post struct {
	Id          uint64  `json:"id"`
	Author      Author  `json:"author"`
	Content     string  `json:"content"`
	ImageUrl    *string `json:"image_url"`
	PublishedAt string  `json:"published_at"`
	UpdatedAt   *string `json:"updated_at"`
}

type LikedPost struct {
	Post
	Liked      bool         `json:"is_liked"`
	LikesCount uint64       `json:"likes_count"`
	Likes      []likes.Like `json:"likes"`
}

// TODO: getting profile feed with and without likes if it makes sence
type AuthorlessPost struct {
	Id          uint64  `json:"id"`
	Content     string  `json:"content"`
	ImageUrl    *string `json:"image_url"`
	PublishedAt string  `json:"published_at"`
	UpdatedAt   *string `json:"updated_at"`
}

type LikedAuthorlessPost struct {
	AuthorlessPost
	Liked      bool         `json:"is_liked"`
	LikesCount uint64       `json:"likes_count"`
	Likes      []likes.Like `json:"likes"`
}

type Page[T any] struct {
	First   uint64 `json:"first"`
	Current uint64 `json:"current"`
	Last    uint64 `json:"last"`
	Posts   []T    `json:"posts"`
}

type PagedFeed Page[LikedPost]
type PagedProfileFeed Page[LikedAuthorlessPost]

type NewPost struct {
	AuthorId   uint64   `json:"author_id"`
	Content    *string  `json:"content"`
	ImagesUrls []string `json:"images_urls"`
}

func (p NewPost) Validate() ero.Error {
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
	if utf8.RuneCountInString(*p.Content) > 1000 {
		faults = append(faults, fieldFault{
			Field:   "content",
			Message: "too long, must be less than 1000 characters",
		})
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
