package repository_test

import (
	"context"
	"testing"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/P4rz1val22/outreach-tool/internal/repository"
	"github.com/P4rz1val22/outreach-tool/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestThreadRepo_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	threadID := uuid.New()
	expected := db.Thread{ID: threadID, Label: "Community Safety Project"}

	mockQuerier.EXPECT().
		GetThreadByID(gomock.Any(), db.GetThreadByIDParams{ID: threadID, UserID: userID}).
		Return(expected, nil)

	repo := repository.NewThreadRepo(mockQuerier)
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

	repo := repository.NewThreadRepo(mockQuerier)
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

	repo := repository.NewThreadRepo(mockQuerier)
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

	repo := repository.NewThreadRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.ListByContact(ctx, contactID, &activeStatus, &tagID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("got %d threads, want 1", len(result))
	}
}
