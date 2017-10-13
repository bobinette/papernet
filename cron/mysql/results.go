package mysql

import (
	"context"

	"github.com/bobinette/papernet/cron"
)

type ResultsRepository struct {
	driver *Driver
}

func NewResultsRepository(driver *Driver) *ResultsRepository {
	repo := &ResultsRepository{
		driver: driver,
	}
	return repo
}

func (r *ResultsRepository) Insert(ctx context.Context, cronID uint, paper cron.Paper) error {
	dbResult := newSearchResult(cronID, paper)
	return r.driver.db.Save(&dbResult).Error
}

func (r *ResultsRepository) GetLastResult(ctx context.Context, cronID uint, source string) (cron.Paper, error) {
	var dbResult SearchResult
	err := r.driver.db.
		Where("cron_id = ?", cronID).
		Where("source = ?", source).
		Order("created_at DESC").
		Limit(1).
		First(&dbResult).
		Error
	if err != nil {
		return cron.Paper{}, err
	}

	return cron.Paper(*dbResult.Result), nil
}
