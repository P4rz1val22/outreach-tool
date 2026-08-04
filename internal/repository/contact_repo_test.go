package repository_test

import (
	"context"
	"errors"
	"testing"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/P4rz1val22/outreach-tool/internal/repository"
	"github.com/P4rz1val22/outreach-tool/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	// If List() reaches the mock at all, the test should fail

	repo := repository.NewContactRepo(mockQuerier)
	_, err := repo.List(context.Background()) // no WithUserID wrapping

	if err == nil {
		t.Error("expected an error when no user is in context, got nil")
	}
}

func TestContactRepo_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	expected := db.Contact{Name: "Anirudh Kashyap"}

	mockQuerier.EXPECT().
		CreateContact(gomock.Any(), db.CreateContactParams{
			Name:   "Anirudh Kashyap",
			Role:   pgtype.Text{String: "DAB Chair", Valid: true},
			UserID: userID,
		}).
		Return(expected, nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)
	role := "DAB Chair"

	result, err := repo.Create(ctx, "Anirudh Kashyap", &role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Anirudh Kashyap" {
		t.Errorf("got %+v, want contact named Anirudh Kashyap", result)
	}
}

func TestContactRepo_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	expected := db.Contact{ID: contactID, Name: "Grace Lyons"}

	mockQuerier.EXPECT().
		UpdateContact(gomock.Any(), db.UpdateContactParams{
			Name:   "Grace Lyons",
			Role:   pgtype.Text{Valid: false},
			ID:     contactID,
			UserID: userID,
		}).
		Return(expected, nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.Update(ctx, contactID, "Grace Lyons", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Grace Lyons" {
		t.Errorf("got %+v, want contact named Grace Lyons", result)
	}
}

func TestContactRepo_Archive(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()

	mockQuerier.EXPECT().
		ArchiveContact(gomock.Any(), db.ArchiveContactParams{ID: contactID, UserID: userID}).
		Return(nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	if err := repo.Archive(ctx, contactID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContactRepo_AddMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	expectedContact := db.Contact{ID: contactID}
	expectedMethod := db.ContactMethod{Type: "email", Value: "grace@example.com"}

	// AddMethod delegates through GetByID first — both calls need mocking
	mockQuerier.EXPECT().
		GetContactByID(gomock.Any(), db.GetContactByIDParams{ID: contactID, UserID: userID}).
		Return(expectedContact, nil)

	mockQuerier.EXPECT().
		CreateContactMethod(gomock.Any(), db.CreateContactMethodParams{
			ContactID: contactID,
			Type:      "email",
			Value:     "grace@example.com",
		}).
		Return(expectedMethod, nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.AddMethod(ctx, contactID, "email", "grace@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value != "grace@example.com" {
		t.Errorf("got %+v, want method with value grace@example.com", result)
	}
}

func TestContactRepo_AddMethod_OwnershipCheckFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()

	// GetContactByID returns an error (e.g. belongs to someone else) —
	// CreateContactMethod must never be called as a result
	mockQuerier.EXPECT().
		GetContactByID(gomock.Any(), db.GetContactByIDParams{ID: contactID, UserID: userID}).
		Return(db.Contact{}, errors.New("not found"))

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	_, err := repo.AddMethod(ctx, contactID, "email", "someone@else.com")
	if err == nil {
		t.Error("expected an error when ownership check fails, got nil")
	}
}

func TestContactRepo_ListMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	contactID := uuid.New()
	expected := []db.ContactMethod{{Type: "phone", Value: "555-0100"}}

	mockQuerier.EXPECT().
		ListContactMethodsByContact(gomock.Any(), db.ListContactMethodsByContactParams{
			ContactID: contactID,
			UserID:    userID,
		}).
		Return(expected, nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	result, err := repo.ListMethods(ctx, contactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("got %d methods, want 1", len(result))
	}
}

func TestContactRepo_DeleteMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()
	methodID := uuid.New()

	mockQuerier.EXPECT().
		DeleteContactMethod(gomock.Any(), db.DeleteContactMethodParams{ID: methodID, UserID: userID}).
		Return(nil)

	repo := repository.NewContactRepo(mockQuerier)
	ctx := auth.WithUserID(context.Background(), userID)

	if err := repo.DeleteMethod(ctx, methodID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
