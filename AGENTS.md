# Agent Instructions for Telegram Knowledge Base & RAG System

This document provides context and rules for AI agents working on this repository. The project is a Telegram Knowledge Base that parses specific channels, vectorizes content, and provides a RAG-based chat interface via a Telegram Mini App (TMA).

## Project Tech Stack

- **Frontend**: Svelte 5 (Runes), SvelteKit, Telegram Web App SDK.
- **Backend Framework**: PocketBase (Golang extended).
- **Search & Vector Engine**: MeiliSearch (v1.6+ with Vector Search enabled).
- **Search Client**: `github.com/meilisearch/meilisearch-go`.
- **Telegram Protocol**: `github.com/gotd/td` (MTProto client for Go).
- **AI/LLM**: `github.com/sashabaranov/go-openai` (Embeddings & Chat).
- **Database**: SQLite (PocketBase default) + MeiliSearch (Index).

## Codebase Architecture

- `src/`: Frontend logic (Web + TMA).
- `pb/`: Golang Source Code (Backend extensions).
  - `pb/main.go`: Entry point, service orchestration.
  - `pb/pkg/parser`: **(Core)** MTProto client. Listens to _selected_ Channel IDs.
  - `pb/pkg/indexer`: **(Core)** Service that handles:
    - LLM Processing (Summarization/Tagging via GPT-4o-mini).
    - Vectorization (OpenAI Embeddings).
    - Syncing to MeiliSearch.
  - `pb/pkg/rag`: Logic for Chat generation (Context retrieval + Answer generation).
  - `pb/migrations`: Database schema changes.
- `session.json`: Telegram session (Gitignored).

## Data Schema & Abstractions

The system revolves around these 4 core collections/entities:

1.  **`chunks`**: The atoms of knowledge.
    - _Source_: Parsed from Telegram messages.
    - _Fields_: `content` (text), `summary` (LLM generated), `embedding` (vector), `meta` (likes, views, date), `source_link` (t.me/...), `channel_id`.
    - _Storage_: Persisted in PocketBase, Indexed in MeiliSearch.
2.  **`chats`**: A conversational session between a User and the AI.
    - _Fields_: `title`, `user_id`, `created`, `updated`.
3.  **`messages`**: History of the chat.
    - _Fields_: `chat_id`, `role` (user/assistant), `content`.
    - _Special Field_: `citations` (JSON array of `chunk` IDs used to generate this answer).
4.  **`users`**: Standard PB users (auth via Telegram).

## Development Rules & Pipeline

### 1. The Ingestion Pipeline (Golang)

The flow must be implemented within `pb/pkg/` to ensure performance and concurrency:

1.  **Filter**: `parser` receives a message. Check if `channel_id` is in the `whitelist`.
2.  **Pre-process**: Extract text, remove spam/ads. Capture metadata (reactions count).
3.  **Enrich (LLM)**: Send text to "Nano" model (e.g., GPT-4o-mini).
    - _Task_: Fix formatting, extract tags, generate a clean summary.
4.  **Vectorize**: Send enriched text to `text-embedding-3-small`.
5.  **Index**:
    - Save raw data to PocketBase `chunks` collection.
    - Push JSON + Vector to MeiliSearch via `indexer` service.

### 2. The RAG Chat Logic

When a user asks a question in the frontend:

1.  **Embed**: Vectorize the user query.
2.  **Search**: Query MeiliSearch using Hybrid Search (Vector + Keyword) to find relevant `chunks`.
3.  **Generate**: Send `System Prompt + Context (Chunks) + User Query` to LLM.
4.  **Track**: Save the resulting message to PB `messages` and link the source chunks in the `citations` field.

### 3. Telegram Safety

- The parser uses `gotd`.
- **Strict Rule**: Only read from channels explicitly defined in the config/whitelist.
- Handle `FLOOD_WAIT` automatically.

### 4. Frontend (TWA)

- The UI works inside Telegram.
- Use `Telegram.WebApp` API for theme params and closing the app.
- Chat interface must support rendering Markdown and "Sources" (clickable chips linking to original TG messages).

## Key Files to Watch

- `pb/pkg/indexer/service.go`: The `Indexer` struct methods (`Vectorize`, `AddToMeili`).
- `pb/pkg/parser/listener.go`: `OnNewMessage` handler with the channel whitelist filter.
- `pb/pkg/rag/engine.go`: The logic that assembles the prompt from MeiliSearch results.
- `src/routes/chat/+page.svelte`: The main chat interface.

## Environment Variables

- `TG_API_ID` / `TG_API_HASH`: Telegram Client.
- `OPENAI_API_KEY`: For Embeddings and Chat.
- `MEILI_HOST` / `MEILI_API_KEY`: Search Engine connection.
- `TARGET_CHANNEL_IDS`: Comma-separated list of Channel IDs to parse.
