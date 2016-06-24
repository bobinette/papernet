package models

type Paper struct {
	ID      int
	Title   []byte
	Read    bool
	Summary []byte

	Authors    []string
	References []int
}
