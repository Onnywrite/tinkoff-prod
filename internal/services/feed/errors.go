package feed

import "errors"

var (
	ErrInternal       = errors.New("internal error")
	ErrAuthorNotFound = errors.New("author not found")
	ErrNoPosts        = errors.New("no posts found")
)
