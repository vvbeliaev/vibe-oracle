package indexer

import (
	"context"
	"fmt"
	"time"

	"svpb-tmpl/pkg/config"

	"github.com/gotd/td/tg"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	IndexName      = "chunks"
	EmbeddingModel = "voyage/voyage-3.5-lite"
	EmbeddingDims  = 1024
)

// ChunkDocument represents a document in the MeiliSearch index.
type ChunkDocument struct {
	ID           string               `json:"id"`
	Content      string               `json:"content"`
	ChannelID    string               `json:"channelId"`
	Link         string               `json:"link"`
	Created      time.Time            `json:"created"`
	Updated      time.Time            `json:"updated"`
	Vectors      map[string][]float32 `json:"_vectors"` // MeiliSearch 1.6+ expects a map if embedders are named
	RankingScore float64              `json:"_rankingScore,omitempty"`
}

// Service handles message indexing: embedding generation, PocketBase storage, and MeiliSearch sync.
type Service struct {
	app      core.App
	meili    meilisearch.ServiceManager
	openai   *openai.Client
	logger   *zap.Logger
	indexUID string
}

// NewService creates a new indexer service.
func NewService(app core.App, cfg *config.Config, logger *zap.Logger) (*Service, error) {
	// Initialize MeiliSearch client
	meiliClient := meilisearch.New(cfg.MeiliHost, meilisearch.WithAPIKey(cfg.MeiliMasterKey))

	// Initialize OpenAI client
	openaiConfig := openai.DefaultConfig(cfg.OpenAIAPIKey)
	if cfg.OpenAIBaseURL != "" {
		openaiConfig.BaseURL = cfg.OpenAIBaseURL
	}
	openaiClient := openai.NewClientWithConfig(openaiConfig)

	svc := &Service{
		app:      app,
		meili:    meiliClient,
		openai:   openaiClient,
		logger:   logger,
		indexUID: IndexName,
	}

	return svc, nil
}

// EnsureIndex creates or updates the MeiliSearch index with proper settings.
func (s *Service) EnsureIndex(ctx context.Context) error {
	// Create index if it doesn't exist
	_, err := s.meili.CreateIndex(&meilisearch.IndexConfig{
		Uid:        s.indexUID,
		PrimaryKey: "id",
	})
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	index := s.meili.Index(s.indexUID)

	// Configure searchable attributes
	searchableAttrs := []string{"content"}
	_, err = index.UpdateSearchableAttributes(&searchableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update searchable attributes: %w", err)
	}

	// Configure filterable attributes
	filterableAttrs := []interface{}{"channelId", "created", "updated"}
	_, err = index.UpdateFilterableAttributes(&filterableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update filterable attributes: %w", err)
	}

	// Configure sortable attributes
	sortableAttrs := []string{"created", "updated"}
	_, err = index.UpdateSortableAttributes(&sortableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update sortable attributes: %w", err)
	}

	// Configure embedders for vector search
	embedders := map[string]meilisearch.Embedder{
		"default": {
			Source:     meilisearch.UserProvidedEmbedderSource,
			Dimensions: EmbeddingDims,
		},
	}
	_, err = index.UpdateEmbedders(embedders)
	if err != nil {
		return fmt.Errorf("failed to update embedders: %w", err)
	}

	s.logger.Info("MeiliSearch index configured", zap.String("index", s.indexUID))
	return nil
}

// IndexMessage processes a Telegram message: generates embedding, saves to PocketBase, and indexes in MeiliSearch.
func (s *Service) IndexMessage(ctx context.Context, msg *tg.Message, channelID int64) error {
	text := msg.Message
	if text == "" {
		return nil
	}

	// Generate embedding
	embedding, err := s.generateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Build source link
	link := fmt.Sprintf("https://t.me/c/%d/%d", channelID, msg.ID)

	// Save to PocketBase
	record, err := s.saveToPocketBase(ctx, text, channelID, link, msg)
	if err != nil {
		return fmt.Errorf("failed to save to PocketBase: %w", err)
	}

	// Index in MeiliSearch
	doc := ChunkDocument{
		ID:        record.Id,
		Content:   text,
		ChannelID: fmt.Sprintf("%d", channelID),
		Link:      link,
		Created:   record.GetDateTime("created").Time(),
		Updated:   record.GetDateTime("updated").Time(),
		Vectors: map[string][]float32{
			"default": embedding,
		},
	}

	if err := s.indexInMeiliSearch(ctx, doc); err != nil {
		return fmt.Errorf("failed to index in MeiliSearch: %w", err)
	}

	s.logger.Info("Message indexed successfully",
		zap.String("id", record.Id),
		zap.Int64("channelId", channelID),
		zap.Int("msgId", msg.ID),
	)

	return nil
}

// generateEmbedding creates a vector embedding for the given text using OpenAI.
func (s *Service) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := s.openai.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(EmbeddingModel),
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return resp.Data[0].Embedding, nil
}

// saveToPocketBase saves the message to the chunks collection.
func (s *Service) saveToPocketBase(_ context.Context, text string, channelID int64, link string, rawMsg *tg.Message) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId("chunks")
	if err != nil {
		return nil, fmt.Errorf("chunks collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("content", text)
	record.Set("channelId", fmt.Sprintf("%d", channelID))
	record.Set("link", link)
	record.Set("raw", rawMsg)

	if err := s.app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

// indexInMeiliSearch adds or updates a document in the search index.
func (s *Service) indexInMeiliSearch(ctx context.Context, doc ChunkDocument) error {
	index := s.meili.Index(s.indexUID)
	primaryKey := "id"
	task, err := index.AddDocuments([]ChunkDocument{doc}, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return err
	}

	return s.waitForTask(ctx, task.TaskUID)
}

// SearchHybrid performs a hybrid search (keyword + vector) in MeiliSearch.
func (s *Service) SearchHybrid(ctx context.Context, query string, queryEmbedding []float32, limit int64) ([]ChunkDocument, error) {
	index := s.meili.Index(s.indexUID)

	searchRes, err := index.Search(query, &meilisearch.SearchRequest{
		Limit: limit,
		Hybrid: &meilisearch.SearchRequestHybrid{
			SemanticRatio: 0.6, // 60% vector, 40% keyword
			Embedder:      "default",
		},
		Vector:           queryEmbedding,
		ShowRankingScore: true,
		RankingScoreThreshold: 0.5,
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert hits to ChunkDocument using DecodeInto
	docs := make([]ChunkDocument, 0, len(searchRes.Hits))
	for _, hit := range searchRes.Hits {
		var doc ChunkDocument
		if err := hit.DecodeInto(&doc); err != nil {
			s.logger.Warn("Failed to decode hit", zap.Error(err))
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// GenerateEmbedding is a public wrapper for generating embeddings (used by RAG service).
func (s *Service) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return s.generateEmbedding(ctx, text)
}

// waitForTask blocks until the task is finished and returns an error if it failed.
func (s *Service) waitForTask(_ context.Context, taskUID int64) error {
	task, err := s.meili.WaitForTask(taskUID, time.Second*10)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed {
		return fmt.Errorf("meilisearch task failed: %s (code: %s, type: %s)", task.Error.Message, task.Error.Code, task.Error.Type)
	}
	return nil
}
