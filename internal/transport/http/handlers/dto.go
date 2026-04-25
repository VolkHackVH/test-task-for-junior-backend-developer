package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Status      taskdomain.Status     `json:"status"`
	Recurrence  *recurrencePayloadDTO `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID           int64                  `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Status       taskdomain.Status      `json:"status"`
	ScheduledFor *time.Time             `json:"scheduled_for,omitempty"`
	SourceTaskID *int64                 `json:"source_task_id,omitempty"`
	Recurrence   *recurrenceResponseDTO `json:"recurrence,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type recurrencePayloadDTO struct {
	RecurrenceType taskdomain.RecurrenceType `json:"recurrence_type"`
	IntervalDays   *int                      `json:"interval_days,omitempty"`
	DayOfMonth     *int                      `json:"day_of_month,omitempty"`
	StartDate      string                    `json:"start_date"`
	IsActive       bool                      `json:"is_active"`
	SpecificDates  []string                  `json:"specific_dates,omitempty"`
}

type recurrenceResponseDTO struct {
	ID             int64                     `json:"id"`
	TaskID         int64                     `json:"task_id"`
	RecurrenceType taskdomain.RecurrenceType `json:"recurrence_type"`
	IntervalDays   *int                      `json:"interval_days,omitempty"`
	DayOfMonth     *int                      `json:"day_of_month,omitempty"`
	StartDate      time.Time                 `json:"start_date"`
	IsActive       bool                      `json:"is_active"`
	SpecificDates  []time.Time               `json:"specific_dates,omitempty"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task, recurrence *taskdomain.Recurrence) taskDTO {
	var recurrenceDTO *recurrenceResponseDTO
	if recurrence != nil {
		dto := newRecurrenceDTO(recurrence)
		recurrenceDTO = &dto
	}

	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		ScheduledFor: task.ScheduledFor,
		SourceTaskID: task.SourceTaskID,
		Recurrence:   recurrenceDTO,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

func newRecurrenceDTO(recurrence *taskdomain.Recurrence) recurrenceResponseDTO {
	return recurrenceResponseDTO{
		ID:             recurrence.ID,
		TaskID:         recurrence.TaskID,
		RecurrenceType: recurrence.RecurrenceType,
		IntervalDays:   recurrence.IntervalDays,
		DayOfMonth:     recurrence.DayOfMonth,
		StartDate:      recurrence.StartDate,
		IsActive:       recurrence.IsActive,
		SpecificDates:  recurrence.SpecificDates,
		CreatedAt:      recurrence.CreatedAt,
		UpdatedAt:      recurrence.UpdatedAt,
	}
}
