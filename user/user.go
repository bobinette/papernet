package user

type User struct {
	Name string `json:"name"`

	Bookmarks []int `json:"bookmarks"`
}
