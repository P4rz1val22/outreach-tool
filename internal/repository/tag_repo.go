package repository

import (
	"context"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/P4rz1val22/outreach-tool/internal/auth"
	"github.com/google/uuid"
)

type TagRepo struct {
	q db.Querier
}

func NewTagRepo(q db.Querier) *TagRepo {
	return &TagRepo{q: q}
}

func (r *TagRepo) Create(ctx context.Context, name string) (db.Tag, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return db.Tag{}, err
	}
	return r.q.CreateTag(ctx, db.CreateTagParams{Name: name, UserID: userID})
}

func (r *TagRepo) List(ctx context.Context) ([]db.Tag, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.ListTags(ctx, userID)
}

func (r *TagRepo) AddToContact(ctx context.Context, contactID, tagID uuid.UUID) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	return r.q.AddTagToContact(ctx, db.AddTagToContactParams{
		ContactID: contactID,
		TagID:     tagID,
		UserID:    userID,
	})
}

func (r *TagRepo) RemoveFromContact(ctx context.Context, contactID, tagID uuid.UUID) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	return r.q.RemoveTagFromContact(ctx, db.RemoveTagFromContactParams{
		ContactID: contactID,
		TagID:     tagID,
		UserID:    userID,
	})
}

func (r *TagRepo) ListForContact(ctx context.Context, contactID uuid.UUID) ([]db.Tag, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.ListTagsForContact(ctx, db.ListTagsForContactParams{
		ContactID: contactID,
		UserID:    userID,
	})
}

func (r *TagRepo) AddToThread(ctx context.Context, threadID, tagID uuid.UUID) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	return r.q.AddTagToThread(ctx, db.AddTagToThreadParams{
		ThreadID: threadID,
		TagID:    tagID,
		UserID:   userID,
	})
}

func (r *TagRepo) RemoveFromThread(ctx context.Context, threadID, tagID uuid.UUID) error {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	return r.q.RemoveTagFromThread(ctx, db.RemoveTagFromThreadParams{
		ThreadID: threadID,
		TagID:    tagID,
		UserID:   userID,
	})
}

func (r *TagRepo) ListForThread(ctx context.Context, threadID uuid.UUID) ([]db.Tag, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.ListTagsForThread(ctx, db.ListTagsForThreadParams{
		ThreadID: threadID,
		UserID:   userID,
	})
}
