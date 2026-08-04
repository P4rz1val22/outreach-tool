package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/google/uuid"
)

type ThreadRepo struct {
	q db.Querier
}

func NewThreadRepo(q db.Querier) *ThreadRepo {
	return &ThreadRepo{q: q}
}

func (r *ThreadRepo) GetByID(ctx context.Context, id uuid.UUID) (db.Thread, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.Thread{}, err
	}
	return r.q.GetThreadByID(ctx, db.GetThreadByIDParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *ThreadRepo) ListByContact(ctx context.Context, contactID uuid.UUID, status *db.ThreadStatus, tagID *uuid.UUID) ([]db.Thread, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	nullStatus := db.NullThreadStatus{Valid: false}
	if status != nil {
		nullStatus = db.NullThreadStatus{ThreadStatus: *status, Valid: true}
	}

	return r.q.ListThreadsByContact(ctx, db.ListThreadsByContactParams{
		ContactID: contactID,
		UserID:    userID,
		Status:    nullStatus,
		TagID:     tagID,
	})
}
