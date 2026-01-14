<script lang="ts">
	import { page } from '$app/state';
	import { messagesStore, chatApi, Messages, MessageControls } from '$lib/apps/chat';
	import { userStore } from '$lib/apps/user';
	import { MessagesRoleOptions } from '$lib';

	const chatId = $derived(page.params.chatId);
	const messages = $derived(messagesStore.messages);
	const user = $derived(userStore.user);

	const userSender = $derived({
		id: user?.id || 'user',
		name: user?.name || 'User',
		avatar: userStore.avatarUrl || '',
		role: 'user'
	});

	const aiSender = {
		id: 'ai',
		name: 'Vibe Oracul',
		avatar: '/logo.svg',
		role: 'ai'
	};

	async function handleSendMessage(content: string) {
		if (!chatId) return;

		try {
			await chatApi.sendMessageSync({
				chat: chatId,
				content: content,
				role: MessagesRoleOptions.user
			});
		} catch (e) {
			console.error('Failed to send message:', e);
		}
	}
</script>

<div class="flex h-full flex-col">
	<!-- Messages Area -->
	<div class="flex-1 overflow-hidden">
		<Messages
			{messages}
			{userSender}
			{aiSender}
			hasMore={messagesStore.page < messagesStore.totalPages}
			loading={messagesStore.loading}
			onLoadMore={() => messagesStore.loadNextPage()}
		/>
	</div>

	<!-- Input Area -->
	<div class="p-4 pb-8 md:pb-6">
		<MessageControls {chatId} {messages} onSend={handleSendMessage} />
	</div>
</div>
