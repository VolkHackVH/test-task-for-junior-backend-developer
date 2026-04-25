package postgres_test

import (
	"context"
	taskdomain "example.com/taskservice/internal/domain/task"
	infraPg "example.com/taskservice/internal/infrastructure/postgres"
	repoPg "example.com/taskservice/internal/repository/postgres"
	"os"
	"testing"
	"time"
)

func TestTaskRepo_Create(t *testing.T) {
	ctx := context.Background()
	dsn := os.Getenv("TEST_DATABASE_DSN")

	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run integration tests")
	}

	pool, err := infraPg.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `TRUNCATE TABLE specific_dates, task_recurrences, tasks RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}

	repo := repoPg.New(pool)

	now := time.Now().UTC()
	task, err := repo.Create(ctx, &taskdomain.Task{
		Title:       "test title",
		Description: "test description",
		Status:      "new",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatal(err)
	}

	startDate := time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)
	date1 := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)

	created, err := repo.RecurrenceCreate(ctx, &taskdomain.Recurrence{
		TaskID:         task.ID,
		RecurrenceType: taskdomain.RecurrenceSpecificDates,
		StartDate:      startDate,
		IsActive:       true,
	}, []time.Time{date1, date2})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.RecurrenceGetByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.TaskID != task.ID {
		t.Errorf("got %v, want %v", got.TaskID, task.ID)
	}
	if got.RecurrenceType != taskdomain.RecurrenceSpecificDates {
		t.Errorf("got %v, want %v", got.RecurrenceType, taskdomain.RecurrenceSpecificDates)
	}
	if len(got.SpecificDates) != 2 {
		t.Fatalf("specific_dates len mismatch: got %d want 2", len(got.SpecificDates))
	}

	got0 := got.SpecificDates[0].UTC().Format("2006-01-02")
	got1 := got.SpecificDates[1].UTC().Format("2006-01-02")
	if got0 != "2026-04-25" || got1 != "2026-05-03" {
		t.Errorf("got %v, want %v", got, "2026-04-25")
	}
}
