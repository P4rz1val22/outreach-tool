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

func TestContactRepo_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	expected := []db.Contact{{Name: "Grace Lyons"}}

	mockQuerier.EXPECT().
		ListContacts(gomock.Any(), userID).
		Return(expected, nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "Grace Lyons" {
		t.Errorf("got %+v, want one contact named Grace Lyons", result)
	}
}

func TestContactRepo_List_NoUserInContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)
	// Deliberately no .EXPECT() call — if List() reaches the mock at all, the test should fail

	repo := repository.NewContactRepo(mockQuerier)
	_, err := repo.List(context.Background()) // no WithUserID wrapping

	if err == nil {
		t.Error("expected an error when no user is in context, got nil")
	}
}
