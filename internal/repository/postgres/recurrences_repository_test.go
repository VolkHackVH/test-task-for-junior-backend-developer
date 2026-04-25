package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	infraPg "example.com/taskservice/internal/infrastructure/postgres"
	repoPg "example.com/taskservice/internal/repository/postgres"
)

func openTestRepository(t *testing.T) (*repoPg.Repository, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run integration tests")
	}

	pool, err := infraPg.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		pool.Close()
	}

	return repoPg.New(pool), cleanup
}

func TestRecurrenceCreate_SpecificDates(t *testing.T) {
	ctx := context.Background()
	repo, closeFn := openTestRepository(t)
	defer closeFn()

	pool, err := infraPg.Open(ctx, os.Getenv("TEST_DATABASE_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `TRUNCATE TABLE specific_dates, task_recurrences, tasks RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	task, err := repo.Create(ctx, &taskdomain.Task{
		Title:       "source task",
		Description: "for recurrence test",
		Status:      taskdomain.StatusNew,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
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
		t.Fatalf("task_id mismatch: got %d want %d", got.TaskID, task.ID)
	}

	if got.RecurrenceType != taskdomain.RecurrenceSpecificDates {
		t.Fatalf("recurrence_type mismatch: got %s", got.RecurrenceType)
	}

	if len(got.SpecificDates) != 2 {
		t.Fatalf("specific_dates len mismatch: got %d want 2", len(got.SpecificDates))
	}

	got0 := got.SpecificDates[0].UTC().Format("2006-01-02")
	got1 := got.SpecificDates[1].UTC().Format("2006-01-02")
	if got0 != "2026-04-25" || got1 != "2026-05-03" {
		t.Fatalf("specific_dates mismatch: got [%s, %s]", got0, got1)
	}
}
