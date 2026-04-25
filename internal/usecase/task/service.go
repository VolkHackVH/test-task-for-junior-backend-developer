package task

import (
	"context"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

const defaultOccurrenceHorizonDays = 30

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) CreateRecurrence(ctx context.Context, input CreateRecurrenceInput) (*taskdomain.Recurrence, error) {
	normalized, err := validateCreateRecurrenceInput(input)
	if err != nil {
		return nil, err
	}

	sourceTask, err := s.repo.GetByID(ctx, normalized.TaskID)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Recurrence{
		TaskID:         normalized.TaskID,
		RecurrenceType: normalized.RecurrenceType,
		IntervalDays:   normalized.IntervalDays,
		DayOfMonth:     normalized.DayOfMonth,
		StartDate:      normalized.StartDate,
		IsActive:       normalized.IsActive,
		CreatedAt:      s.now(),
		UpdatedAt:      s.now(),
	}

	created, err := s.repo.RecurrenceCreate(ctx, model, normalized.SpecificDates)
	if err != nil {
		return nil, err
	}

	if err := s.regenerateFutureOccurrences(ctx, sourceTask, created); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) RecurrenceGetByID(ctx context.Context, id int64) (*taskdomain.Recurrence, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.RecurrenceGetByID(ctx, id)
}

func (s *Service) RecurrenceGetByTaskID(ctx context.Context, id int64) (*taskdomain.Recurrence, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.RecurrenceGetByTaskID(ctx, id)
}

func (s *Service) RecurrenceUpdateByID(ctx context.Context, id int64, input UpdateRecurrenceInput) (*taskdomain.Recurrence, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateRecurrenceInput(input)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.RecurrenceGetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Recurrence{
		ID:             id,
		TaskID:         existing.TaskID,
		RecurrenceType: normalized.RecurrenceType,
		IntervalDays:   normalized.IntervalDays,
		DayOfMonth:     normalized.DayOfMonth,
		StartDate:      normalized.StartDate,
		IsActive:       normalized.IsActive,
		UpdatedAt:      s.now(),
	}

	updated, err := s.repo.RecurrenceUpdateByID(ctx, model, normalized.SpecificDates)
	if err != nil {
		return nil, err
	}

	sourceTask, err := s.repo.GetByID(ctx, updated.TaskID)
	if err != nil {
		return nil, err
	}

	if err := s.regenerateFutureOccurrences(ctx, sourceTask, updated); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) RecurrenceDeleteByID(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.RecurrenceDeleteByID(ctx, id)
}

func (s *Service) RecurrenceListAll(ctx context.Context) ([]*taskdomain.Recurrence, error) {
	return s.repo.RecurrenceListAll(ctx)
}

func (s *Service) regenerateFutureOccurrences(ctx context.Context, sourceTask *taskdomain.Task, recurrence *taskdomain.Recurrence) error {
	now := dateOnlyUTC(s.now())

	if err := s.repo.DeleteFutureOccurrencesBySourceTaskID(ctx, sourceTask.ID, now); err != nil {
		return err
	}

	if !recurrence.IsActive {
		return nil
	}

	horizon := now.AddDate(0, 0, defaultOccurrenceHorizonDays)
	dates := buildOccurrenceDates(recurrence, now, horizon)
	for _, date := range dates {
		scheduledFor := date
		sourceTaskID := sourceTask.ID
		task := &taskdomain.Task{
			Title:        sourceTask.Title,
			Description:  sourceTask.Description,
			Status:       sourceTask.Status,
			ScheduledFor: &scheduledFor,
			SourceTaskID: &sourceTaskID,
			CreatedAt:    s.now(),
			UpdatedAt:    s.now(),
		}

		if err := s.repo.CreateOccurrence(ctx, task); err != nil {
			return err
		}
	}

	return nil
}
