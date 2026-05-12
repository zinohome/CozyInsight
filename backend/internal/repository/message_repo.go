package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	query := `INSERT INTO messages (user_id, title, content, type, is_read)
			  VALUES (:user_id, :title, :content, :type, :is_read)`
	result, err := r.db.NamedExecContext(ctx, query, msg)
	if err != nil {
		return fmt.Errorf("create message failed: %w", err)
	}
	id, _ := result.LastInsertId()
	msg.ID = uint64(id)
	return nil
}

func (r *MessageRepository) ListByUser(ctx context.Context, userID uint64, unreadOnly bool) ([]model.Message, error) {
	var query string
	var args []interface{}
	if unreadOnly {
		query = `SELECT * FROM messages WHERE user_id = ? AND is_read = 0 ORDER BY created_at DESC`
		args = []interface{}{userID}
	} else {
		query = `SELECT * FROM messages WHERE user_id = ? ORDER BY created_at DESC`
		args = []interface{}{userID}
	}
	var msgs []model.Message
	if err := r.db.SelectContext(ctx, &msgs, query, args...); err != nil {
		return nil, fmt.Errorf("list messages failed: %w", err)
	}
	return msgs, nil
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, id uint64) error {
	query := `UPDATE messages SET is_read = 1 WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("mark as read failed: %w", err)
	}
	return nil
}

func (r *MessageRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	query := `UPDATE messages SET is_read = 1 WHERE user_id = ? AND is_read = 0`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("mark all as read failed: %w", err)
	}
	return nil
}

func (r *MessageRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM messages WHERE user_id = ? AND is_read = 0`
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("count unread failed: %w", err)
	}
	return count, nil
}

func (r *MessageRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM messages WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete message failed: %w", err)
	}
	return nil
}
