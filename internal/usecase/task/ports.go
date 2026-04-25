package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	CreateOccurrence(ctx context.Context, task *taskdomain.Task) error
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	DeleteFutureOccurrencesBySourceTaskID(ctx context.Context, sourceTaskID int64, fromDate time.Time) error

	RecurrenceCreate(ctx context.Context, recur *taskdomain.Recurrence, specificDates []time.Time) (*taskdomain.Recurrence, error)
	RecurrenceGetByID(ctx context.Context, id int64) (*taskdomain.Recurrence, error)
	RecurrenceGetByTaskID(ctx context.Context, id int64) (*taskdomain.Recurrence, error)
	RecurrenceUpdateByID(ctx context.Context, recur *taskdomain.Recurrence, specificDates []time.Time) (*taskdomain.Recurrence, error)
	RecurrenceDeleteByID(ctx context.Context, id int64) error
	RecurrenceListAll(ctx context.Context) ([]*taskdomain.Recurrence, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	CreateRecurrence(ctx context.Context, input CreateRecurrenceInput) (*taskdomain.Recurrence, error)
	RecurrenceGetByID(ctx context.Context, id int64) (*taskdomain.Recurrence, error)
	RecurrenceGetByTaskID(ctx context.Context, id int64) (*taskdomain.Recurrence, error)
	RecurrenceUpdateByID(ctx context.Context, id int64, input UpdateRecurrenceInput) (*taskdomain.Recurrence, error)
	RecurrenceDeleteByID(ctx context.Context, id int64) error
	RecurrenceListAll(ctx context.Context) ([]*taskdomain.Recurrence, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type CreateRecurrenceInput struct {
	TaskID         int64
	RecurrenceType taskdomain.RecurrenceType
	IntervalDays   *int
	DayOfMonth     *int
	StartDate      time.Time
	IsActive       bool
	SpecificDates  []time.Time
}

type UpdateRecurrenceInput struct {
	RecurrenceType taskdomain.RecurrenceType
	IntervalDays   *int
	DayOfMonth     *int
	StartDate      time.Time
	IsActive       bool
	SpecificDates  []time.Time
}
