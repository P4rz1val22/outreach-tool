package repository

import (
	"context"
	"time"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *ThreadRepo) GetCurrentPendingCheckIn(ctx context.Context, threadID uuid.UUID) (db.CheckIn, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.CheckIn{}, err
	}
	return r.q.GetCurrentPendingCheckIn(ctx, db.GetCurrentPendingCheckInParams{
		ThreadID: threadID,
		UserID:   userID,
	})
}

func (r *ThreadRepo) Create(ctx context.Context, contactID uuid.UUID, label string, cadenceIntervalDays *int32) (db.Thread, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.Thread{}, err
	}

	var created db.Thread

	err = r.tx.RunTx(ctx, func(qtx db.Querier) error {
		// Verify the contact actually belongs to this user before attaching anything to it
		if _, err := qtx.GetContactByID(ctx, db.GetContactByIDParams{ID: contactID, UserID: userID}); err != nil {
			return err
		}

		thread, err := qtx.CreateThread(ctx, db.CreateThreadParams{
			ContactID:           contactID,
			Label:               label,
			CadenceIntervalDays: cadenceIntervalDays,
		})
		if err != nil {
			return err
		}
		created = thread

		date := pgtype.Date{Time: time.Now(), Valid: true}
		var deadline pgtype.Date
		if cadenceIntervalDays != nil {
			deadline = addDays(date, int(*cadenceIntervalDays))
		}
		// if one-off, deadline stays zero-value/invalid — matches "NULL only possible if cadence is NULL" from the schema

		_, err = qtx.CreateCheckIn(ctx, db.CreateCheckInParams{
			ThreadID: thread.ID,
			Date:     date,
			Deadline: deadline,
		})
		return err
	})

	return created, err
}

func (r *ThreadRepo) RescheduleCheckIn(ctx context.Context, checkInID uuid.UUID, newDate pgtype.Date) (db.CheckIn, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.CheckIn{}, err
	}

	var rescheduled db.CheckIn

	err = r.tx.RunTx(ctx, func(qtx db.Querier) error {
		thread, err := qtx.GetThreadByCheckInID(ctx, db.GetThreadByCheckInIDParams{
			ID: checkInID, UserID: userID,
		})
		if err != nil {
			return err
		}

		var deadline pgtype.Date
		if thread.CadenceIntervalDays != nil {
			deadline = addDays(newDate, int(*thread.CadenceIntervalDays))
		}

		updated, err := qtx.RescheduleCheckIn(ctx, db.RescheduleCheckInParams{
			ID: checkInID, Date: newDate, Deadline: deadline, UserID: userID,
		})
		if err != nil {
			return err
		}
		rescheduled = updated
		return nil
	})

	return rescheduled, err
}
