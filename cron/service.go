package cron

import (
	"context"
	"fmt"

	"gopkg.in/robfig/cron.v2"

	"github.com/bobinette/papernet/clients/imports"
	"github.com/bobinette/papernet/log"
	"github.com/bobinette/papernet/users"
)

const (
	spec = "0 0 0 * * *" // Daily at midnight
	// spec = "2 * * * * *" // Every 2 minutes. For dev
)

type Service struct {
	repo          Repository
	resultRepo    ResultRepository
	importsClient *imports.Client

	notifierFactory NotifierFactory

	logger log.Logger
}

func NewService(
	repo Repository,
	resultRepo ResultRepository,
	notifierFactory NotifierFactory,
	importsClient *imports.Client,
	logger log.Logger,
) *Service {
	return &Service{
		repo:          repo,
		resultRepo:    resultRepo,
		importsClient: importsClient,

		notifierFactory: notifierFactory,

		logger: logger,
	}
}

func (s *Service) GetForUser(ctx context.Context, userID int) ([]Cron, error) {
	return s.repo.GetForUser(ctx, userID)
}

func (s *Service) Insert(ctx context.Context, cron *Cron) error {
	err := s.repo.Insert(ctx, cron)
	if err != nil {
		return err
	}

	// Run the cron, and store the most recent result as a baseline for
	// future runs
	var res map[string]SearchResults
	err = s.importsClient.Search(ctx, cron.Q, 1, 0, cron.Sources).Decode(&res)
	if err != nil {
		return err
	}

	for _, sr := range res {
		for _, paper := range sr.Papers {
			err := s.resultRepo.Insert(ctx, cron.ID, paper)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) StartCron(ctx context.Context) {
	c := cron.New()
	c.AddFunc(spec, func() {
		if err := s.RunCrons(ctx); err != nil {
			s.logger.Errorf("could not execute crons: %v", err)
		} else {
			s.logger.Print("successfully ran crons")
		}
	})
	c.Start()
}

func (s *Service) RunCrons(ctx context.Context) error {
	crons, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	for _, cron := range crons {
		var res map[string]SearchResults

		userCtx := users.AddToContext(ctx, users.User{ID: cron.UserID})
		err := s.importsClient.Search(userCtx, cron.Q, 10, 0, cron.Sources).Decode(&res)
		if err != nil {
			return err
		}

		notifier, err := s.notifierFactory(cron)
		if err != nil {
			return err
		}

		for source, sr := range res {
			last, err := s.resultRepo.GetLastResult(ctx, cron.ID, source)
			if err != nil {
				return err
			}

			papers := make([]Paper, 0, len(sr.Papers))

			for _, paper := range sr.Papers {
				// check the source to make sure last is not empty (can't compare to nil)
				if last.Source != "" && (last.CreatedAt.After(paper.CreatedAt) || last.CreatedAt.Equal(paper.CreatedAt)) {
					fmt.Println(last.CreatedAt, paper.CreatedAt)
					continue
				}

				if paper.ID != 0 {
					// Paper already imported, nothing to do with it
					continue
				}

				papers = append(papers, paper)
			}

			err = notifier.Notify(ctx, papers)
			if err != nil {
				return err
			}

			for _, paper := range papers {
				err = s.resultRepo.Insert(ctx, cron.ID, paper)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
