import { Collections, pb, type Create, type Update } from '$lib';

import type { MessageChunk, ChatResponse } from './models.ts';
import { messagesStore } from './messages.svelte.ts';
import { env } from '$env/dynamic/public';

class ChatApi {
	// Create new chat
	async create(dto: Create<Collections.Chats>) {
		const chat = await pb.collection(Collections.Chats).create(dto);
		return chat;
	}

	// Update chat
	async update(id: string, dto: Update<Collections.Chats>) {
		const chat = await pb.collection(Collections.Chats).update(id, dto);
		return chat;
	}

	async sendMessage(dto: Create<Collections.Messages>, sourceIds?: string[]) {
		if (!dto.content) throw new Error('Content is required');

		messagesStore.addOptimisticMessage(dto);

		const params = new URLSearchParams({ q: dto.content });
		if (sourceIds?.length) {
			params.set('sourceIds', sourceIds.join(','));
		}

		const es = new EventSource(
			`${env.PUBLIC_PB_URL}/api/chats/${dto.chat}/sse?${params.toString()}`,
			{
				withCredentials: true
			}
		);

		es.addEventListener('chunk', (e) => {
			const chunk = JSON.parse(e.data) as MessageChunk;
			messagesStore.addChunk(chunk);
		});
		es.addEventListener('error', (e) => {
			console.error(e);
			es.close();
		});
		es.addEventListener('done', () => {
			es.close();
		});

		es.onerror = (e) => {
			console.error(e);
			es.close();
		};
	}

	async sendMessageSync(dto: Create<Collections.Messages>, sourceIds?: string[]) {
		if (!dto.content) throw new Error('Content is required');

		messagesStore.addOptimisticMessage(dto);

		const response = await fetch(`${env.PUBLIC_PB_URL}/api/chat`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				chatId: dto.chat,
				message: dto.content,
				sourceIds: sourceIds
			})
		});

		if (!response.ok) {
			const err = await response.json();
			throw new Error(err.message || 'Failed to send message');
		}

		return (await response.json()) as ChatResponse;
	}
}

export const chatApi = new ChatApi();
