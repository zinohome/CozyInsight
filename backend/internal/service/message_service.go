package service

import (
	"context"
	"fmt"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type MessageService struct {
	repo *repository.MessageRepository
}

func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) Create(ctx context.Context, userID uint64, title, content, msgType string) (*model.Message, error) {
	msg := &model.Message{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    msgType,
		IsRead:  0,
	}
	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("create message failed: %w", err)
	}
	return msg, nil
}

func (s *MessageService) List(ctx context.Context, userID uint64, unreadOnly bool) ([]model.Message, error) {
	return s.repo.ListByUser(ctx, userID, unreadOnly)
}

func (s *MessageService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *MessageService) MarkAllAsRead(ctx context.Context, userID uint64) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *MessageService) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *MessageService) Delete(ctx context.Context, id uint64, userID uint64) error {
	return s.repo.Delete(ctx, id)
}
