package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/google/uuid"
)

type ContactRepo struct {
	q db.Querier
}

func NewContactRepo(q db.Querier) *ContactRepo {
	return &ContactRepo{q: q}
}

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
