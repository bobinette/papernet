package database

import (
	"github.com/bobinette/papernet/models"
)

type DB interface {
	Get(...int) ([]*models.Paper, error)
	List() ([]*models.Paper, error)

	Insert(*models.Paper) error
	Update(*models.Paper) error

	Delete(int) error

	Close() error
}

type Search interface {
	Find(string) ([]int, error)
	Index(*models.Paper) error

	Close() error
}
