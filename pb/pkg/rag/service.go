package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"svpb-tmpl/pkg/config"
	"svpb-tmpl/pkg/indexer"

	"github.com/pocketbase/pocketbase/core"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	ChatModel          = "gpt-4o-mini"
	MaxContextDocs     = 20
	RelevanceThreshold = 0.5 // Minimum relevance score for documents to be included in context
)

// ChatRequest represents an incoming chat request.
type ChatRequest struct {
	ChatID    string   `json:"chatId"`
	Message   string   `json:"message"`
	SourceIDs []string `json:"sourceIds"`
}

// ChatResponse represents the response to a chat request.
type ChatResponse struct {
	MessageID string   `json:"messageId"`
	Content   string   `json:"content"`
	Citations []Source `json:"citations"`
}

// Source represents a citation source.
type Source struct {
	ID      string `json:"id"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// Service handles RAG-based chat functionality.
type Service struct {
	app     core.App
	indexer *indexer.Service
	openai  *openai.Client
	logger  *zap.Logger
}

// NewService creates a new RAG service.
func NewService(app core.App, indexerSvc *indexer.Service, cfg *config.Config, logger *zap.Logger) *Service {
	// Initialize OpenAI client
	openaiConfig := openai.DefaultConfig(cfg.OpenAIAPIKey)
	if cfg.OpenAIBaseURL != "" {
		openaiConfig.BaseURL = cfg.OpenAIBaseURL
	}
	openaiClient := openai.NewClientWithConfig(openaiConfig)

	return &Service{
		app:     app,
		indexer: indexerSvc,
		openai:  openaiClient,
		logger:  logger,
	}
}

// HandleChatSSE processes a chat request with Server-Sent Events for streaming.
func (s *Service) HandleChatSSE(e *core.RequestEvent) error {
	chatID := e.Request.PathValue("chatId")
	query := e.Request.URL.Query().Get("q")

	if query == "" {
		return e.BadRequestError("Query is required", nil)
	}

	ctx := e.Request.Context()

	// Generate embedding for the query
	embedding, err := s.generateEmbedding(ctx, query)
	if err != nil {
		s.logger.Error("Failed to generate embedding", zap.Error(err))
		return e.InternalServerError("Failed to process query", err)
	}

	// Search for relevant documents
	docs, err := s.indexer.SearchHybrid(ctx, query, embedding, MaxContextDocs)
	if err != nil {
		s.logger.Error("Failed to search documents", zap.Error(err))
		return e.InternalServerError("Search failed", err)
	}

	// Build context from documents
	contextText, sources := s.buildContext(docs)

	// Save user message (if chat exists)
	if chatID != "" {
		_, _ = s.saveMessage(ctx, chatID, "user", query, nil, "final")
	}

	flusher, ok := e.Response.(http.Flusher)
	if !ok {
		return e.InternalServerError("Streaming not supported", nil)
	}

	// Set headers for SSE
	e.Response.Header().Set("Content-Type", "text/event-stream")
	e.Response.Header().Set("Cache-Control", "no-cache")
	e.Response.Header().Set("Connection", "keep-alive")
	e.Response.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering
	e.Response.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Create streaming request
	systemPrompt := `You are a helpful assistant that answers questions based on the provided context from Telegram channels.

RULES:
1. Answer based ONLY on the provided context. If the context doesn't contain relevant information, say so.
2. Be concise and direct in your answers.
3. If referencing specific information, mention the source number (e.g., [1], [2]).
4. Respond in the same language as the user's question.
5. If the context contains code or technical information, format it properly using markdown.`

	userPrompt := query
	if contextText != "" {
		userPrompt = fmt.Sprintf("Context:\n%s\n\nQuestion: %s", contextText, query)
	}

	stream, err := s.openai.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: ChatModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Stream: true,
	})
	if err != nil {
		return e.InternalServerError("Failed to start stream", err)
	}
	defer stream.Close()

	var fullContent strings.Builder

	// Create a placeholder message in DB for streaming
	var aiMsgRecord *core.Record
	if chatID != "" {
		aiMsgRecord, _ = s.saveMessage(ctx, chatID, "ai", "", map[string]interface{}{"citations": sources}, "streaming")
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			break
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			fullContent.WriteString(content)

			// Send chunk to client
			chunk := map[string]interface{}{
				"text":  content,
				"msgId": "",
			}
			if aiMsgRecord != nil {
				chunk["msgId"] = aiMsgRecord.Id
			}

			data, _ := json.Marshal(chunk)
			fmt.Fprintf(e.Response, "event: chunk\ndata: %s\n\n", string(data))
			flusher.Flush()
		}
	}

	// Finalize AI message in DB
	if aiMsgRecord != nil {
		aiMsgRecord.Set("content", fullContent.String())
		aiMsgRecord.Set("status", "final")
		_ = s.app.Save(aiMsgRecord)
	}

	// Update chat title and status if it was "New Chat" or empty
	if chatID != "" {
		chat, err := s.app.FindRecordById("chats", chatID)
		if err == nil {
			isNew := chat.GetString("title") == "New Chat" || chat.GetString("title") == ""
			isEmpty := chat.GetString("status") == "empty"

			if isNew || isEmpty {
				if isNew {
					title := query
					if len(title) > 50 {
						title = title[:47] + "..."
					}
					chat.Set("title", title)
				}
				chat.Set("status", "going")
				_ = s.app.Save(chat)
			}
		}
	}

	fmt.Fprintf(e.Response, "event: done\ndata: {}\n\n")
	flusher.Flush()

	return nil
}

// HandleChat processes a chat request via PocketBase API route.
func (s *Service) HandleChat(e *core.RequestEvent) error {
	// Parse request body
	var req ChatRequest
	if err := e.BindBody(&req); err != nil {
		return e.BadRequestError("Invalid request body", err)
	}

	if req.Message == "" {
		return e.BadRequestError("Message is required", nil)
	}

	ctx := e.Request.Context()

	// Get or create chat
	chatID := req.ChatID
	if chatID == "" {
		chat, err := s.createChat(ctx, req.Message)
		if err != nil {
			s.logger.Error("Failed to create chat", zap.Error(err))
			return e.InternalServerError("Failed to create chat", err)
		}
		chatID = chat.Id
	}

	// Save user message
	userMsgRecord, err := s.saveMessage(ctx, chatID, "user", req.Message, nil, "final")
	if err != nil {
		s.logger.Error("Failed to save user message", zap.Error(err))
		return e.InternalServerError("Failed to save message", err)
	}
	s.logger.Debug("User message saved", zap.String("id", userMsgRecord.Id))

	// Generate embedding for the query
	embedding, err := s.generateEmbedding(ctx, req.Message)
	if err != nil {
		s.logger.Error("Failed to generate embedding", zap.Error(err))
		return e.InternalServerError("Failed to process query", err)
	}

	// Search for relevant documents
	docs, err := s.indexer.SearchHybrid(ctx, req.Message, embedding, MaxContextDocs)
	if err != nil {
		s.logger.Error("Failed to search documents", zap.Error(err))
		return e.InternalServerError("Search failed", err)
	}

	// Build context from documents
	contextText, sources := s.buildContext(docs)

	// Update chat title and status if it was "New Chat" or empty
	if chatID != "" {
		chat, err := s.app.FindRecordById("chats", chatID)
		if err == nil {
			isNew := chat.GetString("title") == "New Chat" || chat.GetString("title") == ""
			isEmpty := chat.GetString("status") == "empty"

			if isNew || isEmpty {
				if isNew {
					title := req.Message
					if len(title) > 50 {
						title = title[:47] + "..."
					}
					chat.Set("title", title)
				}
				chat.Set("status", "going")
				_ = s.app.Save(chat)
			}
		}
	}

	// Generate AI response
	aiResponse, err := s.generateResponse(ctx, req.Message, contextText)
	if err != nil {
		s.logger.Error("Failed to generate response", zap.Error(err))
		return e.InternalServerError("Failed to generate response", err)
	}

	// Save AI message with citations
	meta := map[string]interface{}{
		"citations": sources,
	}
	aiMsgRecord, err := s.saveMessage(ctx, chatID, "ai", aiResponse, meta, "final")
	if err != nil {
		s.logger.Error("Failed to save AI message", zap.Error(err))
		return e.InternalServerError("Failed to save response", err)
	}

	// Return response
	return e.JSON(200, ChatResponse{
		MessageID: aiMsgRecord.Id,
		Content:   aiResponse,
		Citations: sources,
	})
}

// createChat creates a new chat record.
func (s *Service) createChat(ctx context.Context, firstMessage string) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId("chats")
	if err != nil {
		return nil, fmt.Errorf("chats collection not found: %w", err)
	}

	// Generate title from first message (truncate if too long)
	title := firstMessage
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	record := core.NewRecord(collection)
	record.Set("title", title)

	if err := s.app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

// saveMessage saves a message to the messages collection.
func (s *Service) saveMessage(_ context.Context, chatID, role, content string, meta map[string]interface{}, status string) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId("messages")
	if err != nil {
		return nil, fmt.Errorf("messages collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("chat", chatID)
	record.Set("role", role)
	record.Set("content", content)
	record.Set("status", status)
	if meta != nil {
		metaJSON, _ := json.Marshal(meta)
		record.Set("meta", string(metaJSON))
	}

	if err := s.app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

// generateEmbedding creates a vector embedding for the given text.
func (s *Service) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return s.indexer.GenerateEmbedding(ctx, text)
}

// buildContext constructs context text and sources from retrieved documents.
func (s *Service) buildContext(docs []indexer.ChunkDocument) (string, []Source) {
	if len(docs) == 0 {
		return "", nil
	}

	var contextParts []string
	sources := make([]Source, 0, len(docs))

	for i, doc := range docs {
		// Add to context
		contextParts = append(contextParts, fmt.Sprintf("[%d] %s", i+1, doc.Content))

		// Create source with snippet
		snippet := doc.Content
		if len(snippet) > 200 {
			snippet = snippet[:197] + "..."
		}

		sources = append(sources, Source{
			ID:      doc.ID,
			Link:    doc.Link,
			Snippet: snippet,
		})
	}

	return strings.Join(contextParts, "\n\n"), sources
}

// generateResponse generates an AI response using the context and user query.
func (s *Service) generateResponse(ctx context.Context, query, contextText string) (string, error) {
	systemPrompt := `You are a helpful assistant that answers questions based on the provided context from Telegram channels.

RULES:
1. Answer based ONLY on the provided context. If the context doesn't contain relevant information, say so.
2. Be concise and direct in your answers.
3. If referencing specific information, mention the source number (e.g., [1], [2]).
4. Respond in the same language as the user's question.
5. If the context contains code or technical information, format it properly using markdown.`

	userPrompt := query
	if contextText != "" {
		userPrompt = fmt.Sprintf("Context:\n%s\n\nQuestion: %s", contextText, query)
	}

	resp, err := s.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: ChatModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1024,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return resp.Choices[0].Message.Content, nil
}
