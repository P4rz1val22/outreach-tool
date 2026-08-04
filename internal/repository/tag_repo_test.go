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

func TestTagRepo_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	expected := db.Tag{Name: "SGA"}

	mockQuerier.EXPECT().
		CreateTag(gomock.Any(), db.CreateTagParams{Name: "SGA", UserID: userID}).
		Return(expected, nil)

	repo := repository.NewTagRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.Create(ctx, "SGA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "SGA" {
		t.Errorf("got %+v, want tag named SGA", result)
	}
}

func TestTagRepo_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	expected := []db.Tag{{Name: "SGA"}, {Name: "Generate"}}

	mockQuerier.EXPECT().
		ListTags(gomock.Any(), userID).
		Return(expected, nil)

	repo := repository.NewTagRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("got %d tags, want 2", len(result))
	}
}

func TestTagRepo_AddToContact(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	tagID := uuid.New()

	mockQuerier.EXPECT().
		AddTagToContact(gomock.Any(), db.AddTagToContactParams{
			ContactID: contactID,
			TagID:     tagID,
			UserID:    userID,
		}).
		Return(nil)

	repo := repository.NewTagRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	if err := repo.AddToContact(ctx, contactID, tagID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
