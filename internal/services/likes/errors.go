package likes

import "errors"

var (
	ErrAlreadyLiked   = errors.New("post has already been liked")
	ErrAlreadyUnliked = errors.New("post has not been liked yet")
	ErrNotFound       = errors.New("user or post not found")
	ErrNoLikes        = errors.New("post has no likes")
	ErrInternal       = errors.New("internal error")
)
