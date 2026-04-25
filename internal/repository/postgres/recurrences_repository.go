package postgres

import (
	"context"
	"errors"
	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/jackc/pgx/v5"
	"time"
)

func (r *Repository) RecurrenceCreate(ctx context.Context, recur *taskdomain.Recurrence, specificDates []time.Time) (*taskdomain.Recurrence, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO task_recurrences (task_id, recurrence_type, interval_days, day_of_month, start_date, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, task_id, recurrence_type, interval_days, day_of_month, start_date, is_active, created_at, updated_at
	`

	row := tx.QueryRow(ctx, query,
		recur.TaskID,
		recur.RecurrenceType,
		recur.IntervalDays,
		recur.DayOfMonth,
		recur.StartDate,
		recur.IsActive,
	)
	created, err := scanRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	if created.RecurrenceType == taskdomain.RecurrenceSpecificDates {
		if err := insertSpecificDates(ctx, tx, created.ID, specificDates); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) RecurrenceGetByID(ctx context.Context, id int64) (*taskdomain.Recurrence, error) {
	const query = `
		SELECT id, task_id, recurrence_type, interval_days, 
		       day_of_month, start_date, is_active, created_at, updated_at
		FROM task_recurrences
		WHERE id = $1
`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	if found.RecurrenceType == taskdomain.RecurrenceSpecificDates {
		dates, err := r.listSpecificDatesByRecurrenceID(ctx, found.ID)
		if err != nil {
			return nil, err
		}

		found.SpecificDates = dates
	}

	return found, err
}

func (r *Repository) RecurrenceGetByTaskID(ctx context.Context, id int64) (*taskdomain.Recurrence, error) {
	const query = `
		SELECT id, task_id, recurrence_type, interval_days,
		       day_of_month, start_date, is_active, created_at, updated_at
		FROM task_recurrences
		WHERE task_id = $1
`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	if found.RecurrenceType == taskdomain.RecurrenceSpecificDates {
		dates, err := r.listSpecificDatesByRecurrenceID(ctx, found.ID)
		if err != nil {
			return nil, err
		}

		found.SpecificDates = dates
	}

	return found, err
}

func (r *Repository) RecurrenceUpdateByID(ctx context.Context, recur *taskdomain.Recurrence, specificDates []time.Time) (*taskdomain.Recurrence, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE task_recurrences
		SET recurrence_type = $1,
		    interval_days = $2,
		    day_of_month = $3,
		    start_date = $4,
		    is_active = $5,
		    updated_at = $6
		WHERE id = $7
		RETURNING id, task_id, recurrence_type, interval_days, 
		    day_of_month, start_date, is_active, created_at, updated_at
`

	row := tx.QueryRow(ctx, query,
		recur.RecurrenceType,
		recur.IntervalDays,
		recur.DayOfMonth,
		recur.StartDate,
		recur.IsActive,
		recur.UpdatedAt,
		recur.ID,
	)
	updated, err := scanRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM specific_dates WHERE recurrence_id = $1`, updated.ID); err != nil {
		return nil, err
	}

	if updated.RecurrenceType == taskdomain.RecurrenceSpecificDates {
		if err := insertSpecificDates(ctx, tx, updated.ID, specificDates); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *Repository) RecurrenceDeleteByID(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_recurrences WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) RecurrenceListAll(ctx context.Context) ([]*taskdomain.Recurrence, error) {
	const query = `SELECT id, task_id, recurrence_type, interval_days, 
       day_of_month, start_date, is_active, created_at, updated_at 
		FROM task_recurrences
		ORDER BY id DESC
`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	recurrences := make([]*taskdomain.Recurrence, 0)
	for rows.Next() {
		recurrence, err := scanRecurrence(rows)
		if err != nil {
			return nil, err
		}

		if recurrence.RecurrenceType == taskdomain.RecurrenceSpecificDates {
			dates, err := r.listSpecificDatesByRecurrenceID(ctx, recurrence.ID)
			if err != nil {
				return nil, err
			}

			recurrence.SpecificDates = dates
		}

		recurrences = append(recurrences, recurrence)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recurrences, nil
}

func (r *Repository) listSpecificDatesByRecurrenceID(ctx context.Context, recurrenceID int64) ([]time.Time, error) {
	const query = `
		SELECT scheduled_date
		FROM specific_dates
		WHERE recurrence_id = $1
		ORDER BY scheduled_date
`

	rows, err := r.pool.Query(ctx, query, recurrenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var scheduledDate time.Time

		if err := rows.Scan(&scheduledDate); err != nil {
			return nil, err
		}

		dates = append(dates, scheduledDate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dates, nil
}

func insertSpecificDates(ctx context.Context, tx pgx.Tx, recurrenceID int64, dates []time.Time) error {
	const query = `
		INSERT INTO specific_dates (recurrence_id, scheduled_date)
		VALUES ($1, $2)
`

	for _, date := range dates {
		if _, err := tx.Exec(ctx, query, recurrenceID, date); err != nil {
			return err
		}
	}

	return nil
}

type recurrenceScanner interface {
	Scan(dest ...interface{}) error
}

func scanRecurrence(scanner recurrenceScanner) (*taskdomain.Recurrence, error) {
	var recurrence taskdomain.Recurrence

	if err := scanner.Scan(
		&recurrence.ID,
		&recurrence.TaskID,
		&recurrence.RecurrenceType,
		&recurrence.IntervalDays,
		&recurrence.DayOfMonth,
		&recurrence.StartDate,
		&recurrence.IsActive,
		&recurrence.CreatedAt,
		&recurrence.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &recurrence, nil
}
