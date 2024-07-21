package likes

type User struct {
	Id       uint64 `json:"id"`
	Name     string `json:"name"`
	Lastname string `json:"surname"`
	Image    string `json:"image"`
}

type Like struct {
	User    User   `json:"user"`
	LikedAt string `json:"liked_at"`
}

type Page[T any] struct {
	First   uint64 `json:"first"`
	Current uint64 `json:"current"`
	Last    uint64 `json:"last"`
	Count   uint64 `json:"count"`
	Likes   []T    `json:"likes"`
}

type PagedLikes Page[Like]
