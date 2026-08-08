package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/P4rz1val22/outreach-tool/internal/repository"
	"github.com/P4rz1val22/outreach-tool/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/mock/gomock"
)

type FakeTxRunner struct {
	Q db.Querier
}

func (r *FakeTxRunner) RunTx(ctx context.Context, fn func(qtx db.Querier) error) error {
	return fn(r.Q)
}

func TestThreadRepo_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	expected := db.Thread{ID: threadID, Label: "Community Safety Project"}

	mockQuerier.EXPECT().
		GetThreadByID(gomock.Any(), db.GetThreadByIDParams{ID: threadID, UserID: userID}).
		Return(expected, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.GetByID(ctx, threadID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Label != "Community Safety Project" {
		t.Errorf("got %+v, want thread labeled Community Safety Project", result)
	}
}

func TestThreadRepo_GetByID_NoUserInContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)
	// No .EXPECT() set up deliberately — the mock should never be called

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	_, err := repo.GetByID(context.Background(), uuid.New())

	if err == nil {
		t.Error("expected an error when no user is in context, got nil")
	}
}

func TestThreadRepo_ListByContact_NoFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	expected := []db.Thread{{Label: "General 1:1"}}

	mockQuerier.EXPECT().
		ListThreadsByContact(gomock.Any(), db.ListThreadsByContactParams{
			ContactID: contactID,
			UserID:    userID,
			Status:    db.NullThreadStatus{Valid: false},
			TagID:     nil,
		}).
		Return(expected, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.ListByContact(ctx, contactID, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("got %d threads, want 1", len(result))
	}
}

func TestThreadRepo_ListByContact_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	tagID := uuid.New()
	activeStatus := db.ThreadStatusActive
	expected := []db.Thread{{Label: "Urgent Project"}}

	mockQuerier.EXPECT().
		ListThreadsByContact(gomock.Any(), db.ListThreadsByContactParams{
			ContactID: contactID,
			UserID:    userID,
			Status:    db.NullThreadStatus{ThreadStatus: activeStatus, Valid: true},
			TagID:     &tagID,
		}).
		Return(expected, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.ListByContact(ctx, contactID, &activeStatus, &tagID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("got %d threads, want 1", len(result))
	}
}

func TestThreadRepo_CompleteCheckIn_Recurring(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	checkInID := uuid.New()
	cadence := int32(14)
	someFixedDate := time.Now()
	deadline := pgtype.Date{Time: someFixedDate, Valid: true}

	mockQuerier.EXPECT().
		ResolveCheckIn(gomock.Any(), db.ResolveCheckInParams{
			ID: checkInID, Status: db.CheckinStatusCompleted, UserID: userID,
		}).
		Return(db.CheckIn{ID: checkInID, ThreadID: threadID, Deadline: deadline}, nil)

	mockQuerier.EXPECT().
		GetThreadByID(gomock.Any(), db.GetThreadByIDParams{ID: threadID, UserID: userID}).
		Return(db.Thread{ID: threadID, CadenceIntervalDays: &cadence}, nil)

	mockQuerier.EXPECT().
		CreateCheckIn(gomock.Any(), gomock.Any()).
		Return(db.CheckIn{}, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	if err := repo.CompleteCheckIn(ctx, checkInID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThreadRepo_CompleteCheckIn_OneOff_Archives(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	checkInID := uuid.New()

	mockQuerier.EXPECT().
		ResolveCheckIn(gomock.Any(), gomock.Any()).
		Return(db.CheckIn{ID: checkInID, ThreadID: threadID}, nil)

	mockQuerier.EXPECT().
		GetThreadByID(gomock.Any(), db.GetThreadByIDParams{ID: threadID, UserID: userID}).
		Return(db.Thread{ID: threadID, CadenceIntervalDays: nil}, nil) // one-off

	// CreateCheckIn must NOT be called — no .EXPECT() for it
	mockQuerier.EXPECT().
		ArchiveThread(gomock.Any(), db.ArchiveThreadParams{ID: threadID, UserID: userID}).
		Return(nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	if err := repo.CompleteCheckIn(ctx, checkInID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThreadRepo_GetCurrentPendingCheckIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	expected := db.CheckIn{ThreadID: threadID, Status: db.CheckinStatusPending}

	mockQuerier.EXPECT().
		GetCurrentPendingCheckIn(gomock.Any(), db.GetCurrentPendingCheckInParams{
			ThreadID: threadID,
			UserID:   userID,
		}).
		Return(expected, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.GetCurrentPendingCheckIn(ctx, threadID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != db.CheckinStatusPending {
		t.Errorf("got status %v, want pending", result.Status)
	}
}

func TestThreadRepo_Create_Recurring(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	cadence := int32(14)

	mockQuerier.EXPECT().
		GetContactByID(gomock.Any(), db.GetContactByIDParams{ID: contactID, UserID: userID}).
		Return(db.Contact{ID: contactID, UserID: userID}, nil)

	mockQuerier.EXPECT().
		CreateThread(gomock.Any(), db.CreateThreadParams{
			ContactID:           contactID,
			Label:               "General 1:1",
			CadenceIntervalDays: &cadence,
		}).
		Return(db.Thread{Label: "General 1:1", CadenceIntervalDays: &cadence}, nil)

	// Date is time.Now()-based, so we can't assert exact params — just confirm it's called
	mockQuerier.EXPECT().
		CreateCheckIn(gomock.Any(), gomock.Any()).
		Return(db.CheckIn{}, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.Create(ctx, contactID, "General 1:1", &cadence)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Label != "General 1:1" {
		t.Errorf("got %+v, want thread labeled General 1:1", result)
	}
}

func TestThreadRepo_Create_OneOff(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()

	mockQuerier.EXPECT().
		GetContactByID(gomock.Any(), db.GetContactByIDParams{ID: contactID, UserID: userID}).
		Return(db.Contact{ID: contactID, UserID: userID}, nil)

	mockQuerier.EXPECT().
		CreateThread(gomock.Any(), db.CreateThreadParams{
			ContactID:           contactID,
			Label:               "Generate outreach",
			CadenceIntervalDays: nil,
		}).
		Return(db.Thread{Label: "Generate outreach", CadenceIntervalDays: nil}, nil)

	mockQuerier.EXPECT().
		CreateCheckIn(gomock.Any(), gomock.Any()).
		Return(db.CheckIn{}, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.Create(ctx, contactID, "Generate outreach", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Label != "Generate outreach" {
		t.Errorf("got %+v, want thread labeled Generate outreach", result)
	}
}

func TestThreadRepo_Create_OwnershipCheckFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()

	mockQuerier.EXPECT().
		GetContactByID(gomock.Any(), db.GetContactByIDParams{ID: contactID, UserID: userID}).
		Return(db.Contact{}, errors.New("not found"))

	// CreateThread and CreateCheckIn must NEVER be called — no .EXPECT() for either

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	_, err := repo.Create(ctx, contactID, "Should not be created", nil)
	if err == nil {
		t.Error("expected an error when contact ownership check fails, got nil")
	}
}

func TestThreadRepo_RescheduleCheckIn_Recurring(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	checkInID := uuid.New()
	cadence := int32(7)
	newDate := pgtype.Date{Time: time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), Valid: true}

	mockQuerier.EXPECT().
		GetThreadByCheckInID(gomock.Any(), db.GetThreadByCheckInIDParams{ID: checkInID, UserID: userID}).
		Return(db.Thread{ID: threadID, CadenceIntervalDays: &cadence}, nil)

	mockQuerier.EXPECT().
		RescheduleCheckIn(gomock.Any(), gomock.Any()).
		Return(db.CheckIn{ID: checkInID, ThreadID: threadID, Date: newDate}, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.RescheduleCheckIn(ctx, checkInID, newDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != checkInID {
		t.Errorf("got %+v, want the rescheduled check-in", result)
	}
}

func TestThreadRepo_RescheduleCheckIn_OneOff_NoDeadlineComputed(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	checkInID := uuid.New()
	newDate := pgtype.Date{Time: time.Now(), Valid: true}

	mockQuerier.EXPECT().
		GetThreadByCheckInID(gomock.Any(), db.GetThreadByCheckInIDParams{ID: checkInID, UserID: userID}).
		Return(db.Thread{ID: threadID, CadenceIntervalDays: nil}, nil)

	// Deadline should come through as the zero-value/invalid pgtype.Date, not a computed one
	mockQuerier.EXPECT().
		RescheduleCheckIn(gomock.Any(), db.RescheduleCheckInParams{
			ID: checkInID, Date: newDate, Deadline: pgtype.Date{}, UserID: userID,
		}).
		Return(db.CheckIn{ID: checkInID}, nil)

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	_, err := repo.RescheduleCheckIn(ctx, checkInID, newDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThreadRepo_RescheduleCheckIn_OwnershipCheckFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	checkInID := uuid.New()

	mockQuerier.EXPECT().
		GetThreadByCheckInID(gomock.Any(), db.GetThreadByCheckInIDParams{ID: checkInID, UserID: userID}).
		Return(db.Thread{}, errors.New("not found"))

	// RescheduleCheckIn must never be called — no .EXPECT() for it

	repo := repository.NewThreadRepo(mockQuerier, &FakeTxRunner{Q: mockQuerier})
	ctx := auth.WithUserID(context.Background(), userID)

	_, err := repo.RescheduleCheckIn(ctx, checkInID, pgtype.Date{Time: time.Now(), Valid: true})
	if err == nil {
		t.Error("expected an error when ownership check fails, got nil")
	}
}
