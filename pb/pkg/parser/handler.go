package parser

import (
	"context"
	"fmt"

	"svpb-tmpl/pkg/config"
	"svpb-tmpl/pkg/indexer"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// Handler processes incoming Telegram messages and indexes them.
type Handler struct {
	cfg     *config.Config
	indexer *indexer.Service
	logger  *zap.Logger
}

// NewHandler creates a new message handler.
func NewHandler(cfg *config.Config, indexerSvc *indexer.Service, logger *zap.Logger) *Handler {
	return &Handler{
		cfg:     cfg,
		indexer: indexerSvc,
		logger:  logger,
	}
}

// HandleMessage processes an incoming Telegram message.
func (h *Handler) HandleMessage(ctx context.Context, msg *tg.Message, chatID int64) error {
	// Skip empty messages
	if msg.Message == "" {
		return nil
	}

	fmt.Printf("New Message from Chat ID: %d\n", chatID) 

	// Check if chat is in whitelist
	if !h.cfg.IsChatAllowed(chatID) {
		h.logger.Debug("Message from non-whitelisted chat, skipping",
			zap.Int64("chatId", chatID),
			zap.Int("msgId", msg.ID),
		)
		return nil
	}

	h.logger.Info("Processing message",
		zap.Int64("chatId", chatID),
		zap.Int("msgId", msg.ID),
		zap.Int("textLength", len(msg.Message)),
	)

	// Index the message
	if err := h.indexer.IndexMessage(ctx, msg, chatID); err != nil {
		h.logger.Error("Failed to index message",
			zap.Error(err),
			zap.Int64("chatId", chatID),
			zap.Int("msgId", msg.ID),
		)
		return nil // Don't propagate error to avoid stopping the listener
	}

	return nil
}
