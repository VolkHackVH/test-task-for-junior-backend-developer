package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceOddDays       RecurrenceType = "odd_days"
	RecurrenceEvenDays      RecurrenceType = "even_days"
)

type Task struct {
	ID           int64      `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       Status     `json:"status"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	SourceTaskID *int64     `json:"source_task_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Recurrence struct {
	ID             int64          `json:"id"`
	TaskID         int64          `json:"task_id"`
	RecurrenceType RecurrenceType `json:"recurrence_type"`
	IntervalDays   *int           `json:"interval_days,omitempty"`
	DayOfMonth     *int           `json:"day_of_month,omitempty"`
	StartDate      time.Time      `json:"start_date"`
	IsActive       bool           `json:"is_active"`
	SpecificDates  []time.Time    `json:"specific_dates,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type RecurrenceSpecificDate struct {
	ID            int64     `json:"id"`
	RecurrenceID  int64     `json:"recurrence_id"`
	ScheduledDate time.Time `json:"scheduled_date"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceOddDays, RecurrenceEvenDays:
		return true
	default:
		return false
	}
}
