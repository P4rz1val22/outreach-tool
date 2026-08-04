package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ContactRepo struct {
	q db.Querier
}

func NewContactRepo(q db.Querier) *ContactRepo {
	return &ContactRepo{q: q}
}

// Helpers

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// Repository Methods

func (r *ContactRepo) List(ctx context.Context) ([]db.Contact, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.ListContacts(ctx, userID)
}

func (r *ContactRepo) GetByID(ctx context.Context, id uuid.UUID) (db.Contact, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.Contact{}, err
	}
	return r.q.GetContactByID(ctx, db.GetContactByIDParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *ContactRepo) Create(ctx context.Context, name string, role *string) (db.Contact, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.Contact{}, err
	}
	return r.q.CreateContact(ctx, db.CreateContactParams{
		Name:   name,
		Role:   toPgText(role),
		UserID: userID,
	})
}

func (r *ContactRepo) Update(ctx context.Context, id uuid.UUID, name string, role *string) (db.Contact, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.Contact{}, err
	}
	return r.q.UpdateContact(ctx, db.UpdateContactParams{
		Name:   name,
		Role:   toPgText(role),
		ID:     id,
		UserID: userID,
	})
}

func (r *ContactRepo) Archive(ctx context.Context, id uuid.UUID) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	return r.q.ArchiveContact(ctx, db.ArchiveContactParams{ID: id, UserID: userID})
}

func (r *ContactRepo) AddMethod(ctx context.Context, contactID uuid.UUID, methodType, value string) (db.ContactMethod, error) {
	if _, err := r.GetByID(ctx, contactID); err != nil {
		return db.ContactMethod{}, err
	}
	return r.q.CreateContactMethod(ctx, db.CreateContactMethodParams{
		ContactID: contactID,
		Type:      methodType,
		Value:     value,
	})
}

func (r *ContactRepo) ListMethods(ctx context.Context, contactID uuid.UUID) ([]db.ContactMethod, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.ListContactMethodsByContact(ctx, db.ListContactMethodsByContactParams{
		ContactID: contactID,
		UserID:    userID,
	})
}

func (r *ContactRepo) DeleteMethod(ctx context.Context, methodID uuid.UUID) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	return r.q.DeleteContactMethod(ctx, db.DeleteContactMethodParams{ID: methodID, UserID: userID})
}
