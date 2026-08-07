package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/google/uuid"
)

type ThreadRepo struct {
	q  db.Querier
	tx TxRunner
}

func NewThreadRepo(q db.Querier, tx TxRunner) *ThreadRepo {
	return &ThreadRepo{q: q, tx: tx}
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

func (r *ThreadRepo) CompleteCheckIn(ctx context.Context, checkInID uuid.UUID) error {
	return r.resolveCheckIn(ctx, checkInID, db.CheckinStatusCompleted)
}

func (r *ThreadRepo) SkipCheckIn(ctx context.Context, checkInID uuid.UUID) error {
	return r.resolveCheckIn(ctx, checkInID, db.CheckinStatusSkipped)
}

func (r *ThreadRepo) resolveCheckIn(ctx context.Context, checkInID uuid.UUID, status db.CheckinStatus) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	return r.tx.RunTx(ctx, func(qtx db.Querier) error {
		resolved, err := qtx.ResolveCheckIn(ctx, db.ResolveCheckInParams{
			ID: checkInID, Status: status, UserID: userID,
		})
		if err != nil {
			return err
		}

		thread, err := qtx.GetThreadByID(ctx, db.GetThreadByIDParams{
			ID: resolved.ThreadID, UserID: userID,
		})
		if err != nil {
			return err
		}

		if thread.CadenceIntervalDays == nil {
			return qtx.ArchiveThread(ctx, db.ArchiveThreadParams{ID: thread.ID, UserID: userID})
		}

		nextDeadline := addDays(resolved.Deadline, int(*thread.CadenceIntervalDays))
		_, err = qtx.CreateCheckIn(ctx, db.CreateCheckInParams{
			ThreadID: thread.ID, Date: resolved.Deadline, Deadline: nextDeadline,
		})
		return err
	})
}
