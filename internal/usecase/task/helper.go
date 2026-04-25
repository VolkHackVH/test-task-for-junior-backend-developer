package task

import (
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func dateOnlyUTC(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func buildOccurrenceDates(recurrence *taskdomain.Recurrence, from, to time.Time) []time.Time {
	if recurrence == nil {
		return nil
	}

	from = dateOnlyUTC(from)
	to = dateOnlyUTC(to)
	start := dateOnlyUTC(recurrence.StartDate)

	if to.Before(from) {
		return nil
	}

	switch recurrence.RecurrenceType {
	case taskdomain.RecurrenceDaily:
		if recurrence.IntervalDays == nil || *recurrence.IntervalDays <= 0 {
			return nil
		}

		interval := *recurrence.IntervalDays
		dates := make([]time.Time, 0)
		begin := maxDate(start, from)
		for date := begin; !date.After(to); date = date.AddDate(0, 0, 1) {
			days := int(date.Sub(start).Hours() / 24)
			if days >= 0 && days%interval == 0 {
				dates = append(dates, date)
			}
		}

		return dates

	case taskdomain.RecurrenceMonthly:
		if recurrence.DayOfMonth == nil {
			return nil
		}

		target := *recurrence.DayOfMonth
		dates := make([]time.Time, 0)
		begin := maxDate(start, from)
		for date := begin; !date.After(to); date = date.AddDate(0, 0, 1) {
			if date.Day() == target {
				dates = append(dates, date)
			}
		}

		return dates

	case taskdomain.RecurrenceSpecificDates:
		dates := make([]time.Time, 0, len(recurrence.SpecificDates))
		for _, date := range recurrence.SpecificDates {
			normalized := dateOnlyUTC(date)
			if normalized.Before(start) || normalized.Before(from) || normalized.After(to) {
				continue
			}
			dates = append(dates, normalized)
		}

		return dates

	case taskdomain.RecurrenceOddDays, taskdomain.RecurrenceEvenDays:
		dates := make([]time.Time, 0)
		begin := maxDate(start, from)
		for date := begin; !date.After(to); date = date.AddDate(0, 0, 1) {
			day := date.Day()
			if recurrence.RecurrenceType == taskdomain.RecurrenceOddDays && day%2 == 1 {
				dates = append(dates, date)
			}
			if recurrence.RecurrenceType == taskdomain.RecurrenceEvenDays && day%2 == 0 {
				dates = append(dates, date)
			}
		}

		return dates
	default:
		return nil
	}
}

type validatedRecurrenceFields struct {
	RecurrenceType taskdomain.RecurrenceType
	IntervalDays   *int
	DayOfMonth     *int
	StartDate      time.Time
	SpecificDates  []time.Time
}

func validateRecurrenceFields(
	recurrenceType taskdomain.RecurrenceType,
	intervalDays *int,
	dayOfMonth *int,
	startDate time.Time,
	specificDates []time.Time,
) (validatedRecurrenceFields, error) {
	recurrenceType = taskdomain.RecurrenceType(strings.TrimSpace(string(recurrenceType)))
	if recurrenceType == "" {
		return validatedRecurrenceFields{}, fmt.Errorf("%w: recurrence_type is required", ErrInvalidInput)
	}

	if !recurrenceType.Valid() {
		return validatedRecurrenceFields{}, fmt.Errorf("%w: invalid recurrence_type", ErrInvalidInput)
	}

	if startDate.IsZero() {
		return validatedRecurrenceFields{}, fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}

	switch recurrenceType {
	case taskdomain.RecurrenceDaily:
		if intervalDays == nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: interval_days is required for daily recurrence", ErrInvalidInput)
		}
		if *intervalDays <= 0 {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: interval_days must be greater than 0", ErrInvalidInput)
		}
		if dayOfMonth != nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: day_of_month must be empty for daily recurrence", ErrInvalidInput)
		}
		if len(specificDates) > 0 {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: specific_dates must be empty for daily recurrence", ErrInvalidInput)
		}

	case taskdomain.RecurrenceMonthly:
		if dayOfMonth == nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: day_of_month is required for monthly recurrence", ErrInvalidInput)
		}
		if *dayOfMonth < 1 || *dayOfMonth > 30 {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}
		if intervalDays != nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: interval_days must be empty for monthly recurrence", ErrInvalidInput)
		}
		if len(specificDates) > 0 {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: specific_dates must be empty for monthly recurrence", ErrInvalidInput)
		}

	case taskdomain.RecurrenceSpecificDates:
		if len(specificDates) == 0 {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: specific_dates is required for specific_dates recurrence", ErrInvalidInput)
		}
		if intervalDays != nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: interval_days must be empty for specific_dates recurrence", ErrInvalidInput)
		}
		if dayOfMonth != nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: day_of_month must be empty for specific_dates recurrence", ErrInvalidInput)
		}

		seen := make(map[string]struct{})
		for _, date := range specificDates {
			if date.IsZero() {
				return validatedRecurrenceFields{}, fmt.Errorf("%w: specific_dates contains empty date", ErrInvalidInput)
			}

			key := date.Format("2006-01-02")
			if _, ok := seen[key]; ok {
				return validatedRecurrenceFields{}, fmt.Errorf("%w: specific_dates contains duplicate date %s", ErrInvalidInput, key)
			}
			seen[key] = struct{}{}
		}

	case taskdomain.RecurrenceOddDays, taskdomain.RecurrenceEvenDays:
		if intervalDays != nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: interval_days must be empty for %s recurrence", ErrInvalidInput, recurrenceType)
		}
		if dayOfMonth != nil {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: day_of_month must be empty for %s recurrence", ErrInvalidInput, recurrenceType)
		}
		if len(specificDates) > 0 {
			return validatedRecurrenceFields{}, fmt.Errorf("%w: specific_dates must be empty for %s recurrence", ErrInvalidInput, recurrenceType)
		}
	}

	return validatedRecurrenceFields{
		RecurrenceType: recurrenceType,
		IntervalDays:   intervalDays,
		DayOfMonth:     dayOfMonth,
		StartDate:      startDate,
		SpecificDates:  specificDates,
	}, nil
}

func validateCreateRecurrenceInput(input CreateRecurrenceInput) (CreateRecurrenceInput, error) {
	if input.TaskID <= 0 {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: task_id must be positive", ErrInvalidInput)
	}

	normalized, err := validateRecurrenceFields(
		input.RecurrenceType,
		input.IntervalDays,
		input.DayOfMonth,
		input.StartDate,
		input.SpecificDates,
	)
	if err != nil {
		return CreateRecurrenceInput{}, err
	}

	input.RecurrenceType = normalized.RecurrenceType
	input.IntervalDays = normalized.IntervalDays
	input.DayOfMonth = normalized.DayOfMonth
	input.StartDate = normalized.StartDate
	input.SpecificDates = normalized.SpecificDates

	return input, nil
}

func validateUpdateRecurrenceInput(input UpdateRecurrenceInput) (UpdateRecurrenceInput, error) {
	normalized, err := validateRecurrenceFields(
		input.RecurrenceType,
		input.IntervalDays,
		input.DayOfMonth,
		input.StartDate,
		input.SpecificDates,
	)
	if err != nil {
		return UpdateRecurrenceInput{}, err
	}

	input.RecurrenceType = normalized.RecurrenceType
	input.IntervalDays = normalized.IntervalDays
	input.DayOfMonth = normalized.DayOfMonth
	input.StartDate = normalized.StartDate
	input.SpecificDates = normalized.SpecificDates

	return input, nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
