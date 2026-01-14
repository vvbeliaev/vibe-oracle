# 🔮 Vibe Oracle

> **Transform your Telegram channels into an intelligent, searchable knowledge base.**

**Vibe Oracul** is a sophisticated RAG (Retrieval-Augmented Generation) system designed specifically for the Telegram ecosystem. It bridges the gap between fragmented channel messages and structured knowledge, providing a seamless AI-powered chat interface directly within a Telegram Mini App.

---

## ✨ Key Features

### 📥 Real-time Knowledge Ingestion

- **MTProto Integration**: Powered by `gotd`, the system listens to whitelisted Telegram channels in real-time.
- **Automated Processing**: Every message is automatically parsed, cleaned, and prepared for indexing without manual intervention.
- **Smart Filtering**: Built-in support for channel whitelisting and spam/ad removal to ensure only high-quality data enters your knowledge base.

### 🧠 Advanced Search & Retrieval

- **Hybrid Search Engine**: Combines the precision of keyword search with the deep understanding of **Vector Search** via MeiliSearch (v1.6+).
- **State-of-the-Art Embeddings**: Uses high-performance embedding models (Voyage AI / OpenAI) to capture the semantic meaning of every message.
- **Contextual Ranking**: Sophisticated relevance scoring ensures the most pertinent information is always used to generate AI responses.

### 💬 RAG-Powered Chat Interface

- **Context-Strict Answers**: The AI (GPT-4o-mini) answers questions based _exclusively_ on your indexed Telegram content, minimizing hallucinations.
- **Streaming UI**: Experience lightning-fast, real-time response generation via Server-Sent Events (SSE).
- **Verifiable Citations**: Every answer includes clickable "Source" chips that link directly back to the original Telegram messages.

### 📱 "Bleeding Edge" Tech Stack

- **Frontend**: Svelte 5 (Runes) + Tailwind 4 + DaisyUI 5 for a hyper-reactive Telegram Mini App experience.
- **Backend**: A single-binary architecture using **PocketBase** extended with custom **Golang** services.
- **Search**: MeiliSearch for sub-millisecond search performance across thousands of messages.

---

## 🚀 Future Potential

Vibe Oracul is built with scalability and extensibility in mind. Its architecture allows for:

- **🌍 Multilingual Intelligence**: Cross-lingual search and response generation, making your knowledge base accessible to a global audience.
- **📈 Knowledge Analytics**: Insights into trending topics, message engagement, and identifying "knowledge gaps" within your channels.
- **💎 Monetization & Paywalls**: Integrated support for premium access tiers, subscriptions, and gated content (built-in Paywall components).
- **🔄 Multi-Source Integration**: Expanding beyond Telegram to index Slack, Discord, and web-based documentation into a unified "Oracle."
- **👥 Collaborative AI**: Tools for users to tag, rate, and refine the knowledge base, creating a self-improving AI ecosystem.

---

## 🛠 Tech Stack Summary

| Layer             | Technology                                                                                                                          |
| :---------------- | :---------------------------------------------------------------------------------------------------------------------------------- |
| **Frontend**      | [Svelte 5](https://svelte.dev), [SvelteKit](https://kit.svelte.dev), [Telegram Web App SDK](https://core.telegram.org/bots/webapps) |
| **Backend**       | [PocketBase](https://pocketbase.io) (Go), [gotd/td](https://github.com/gotd/td) (MTProto)                                           |
| **Search/Vector** | [MeiliSearch](https://www.meilisearch.com) (Vector Search enabled)                                                                  |
| **LLM/AI**        | [OpenAI](https://openai.com) (GPT-4o-mini, Text-Embeddings), [Voyage AI](https://www.voyageai.com)                                  |
| **Database**      | SQLite (via PocketBase)                                                                                                             |

---

## 📜 License

MIT © [Vladimir Beliaev](https://github.com/vvbeliaev)
