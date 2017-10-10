package cron

import (
	"context"
	"time"
)

type Cron struct {
	ID      uint     `json:"id"`
	UserID  int      `json:"userId"`
	Sources []string `json:"sources"`
	Q       string   `json:"q"`
}

type Repository interface {
	GetForUser(ctx context.Context, userID int) ([]Cron, error)
	List(ctx context.Context) ([]Cron, error)
	Insert(ctx context.Context, cron *Cron) error
	Delete(ctx context.Context, id uint) error
}

type Paper struct {
	ID int `json:"id"`

	Source    string `json:"source"`
	Reference string `json:"reference"`

	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Tags       []string `json:"tags"`
	Authors    []string `json:"authors"`
	References []string `json:"references"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SearchResults struct {
	Papers     []Paper `json:"papers"`
	Pagination struct {
		Limit  uint `json:"limit"`
		Offset uint `json:"offset"`
		Total  uint `json:"total"`
	} `json:"pagination"`
}

type ResultRepository interface {
	Insert(ctx context.Context, cronID uint, paper Paper) error
	GetLastResult(ctx context.Context, cronID uint, source string) (Paper, error)
}

type Notifier interface {
	Notify(ctx context.Context, papers []Paper) error
}

type NotifierFactory func(cron Cron) (Notifier, error)
