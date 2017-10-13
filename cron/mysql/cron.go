package mysql

import (
	"context"

	"github.com/bobinette/papernet/cron"
	"github.com/bobinette/papernet/errors"
)

type Repository struct {
	driver *Driver
}

func NewRepository(driver *Driver) *Repository {
	repo := &Repository{
		driver: driver,
	}
	return repo
}

func (r *Repository) List(ctx context.Context) ([]cron.Cron, error) {
	var dbCrons []Cron
	err := r.driver.db.
		Find(&dbCrons).
		Error
	if err != nil {
		return nil, err
	}

	crons := make([]cron.Cron, len(dbCrons))
	for i, dbCron := range dbCrons {
		crons[i] = dbCron.format()
	}
	return crons, nil
}

func (r *Repository) GetForUser(ctx context.Context, userID int) ([]cron.Cron, error) {
	var dbCrons []Cron
	err := r.driver.db.
		Where("user_id = ?", userID).
		Find(&dbCrons).
		Error
	if err != nil {
		return nil, err
	}

	crons := make([]cron.Cron, len(dbCrons))
	for i, dbCron := range dbCrons {
		crons[i] = dbCron.format()
	}
	return crons, nil
}

func (r *Repository) Insert(ctx context.Context, c *cron.Cron) error {
	if c.ID != 0 {
		return errors.New("cannot update cron", errors.BadRequest())
	}

	dbCron := newCron(*c)
	err := r.driver.db.Save(&dbCron).Error
	if err != nil {
		return err
	}

	c.ID = dbCron.ID
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.driver.db.Delete(Cron{ID: id}).Error
}
