<script lang="ts">
	import { Send, Paperclip } from 'lucide-svelte';
	import { page } from '$app/state';
	import { messagesStore, chatApi, Messages } from '$lib/apps/chat';
	import { userStore } from '$lib/apps/user';
	import { Button, Textarea, MessagesRoleOptions } from '$lib';

	const chatId = $derived(page.params.chatId);
	const messages = $derived(messagesStore.messages);
	const user = $derived(userStore.user);

	let input = $state('');
	let loading = $state(false);

	const userSender = $derived({
		id: user?.id || 'user',
		name: user?.name || 'User',
		avatar: userStore.avatarUrl || '',
		role: 'user'
	});

	const aiSender = {
		id: 'ai',
		name: 'Vibe Oracul',
		avatar: '/logo.svg', // Assuming there's a logo
		role: 'ai'
	};

	async function handleSendMessage() {
		if (!input.trim() || !chatId || loading) return;

		const content = input;
		input = '';
		loading = true;

		try {
			await chatApi.sendMessage({
				chat: chatId,
				content: content,
				role: MessagesRoleOptions.user
			});
		} catch (e) {
			console.error('Failed to send message:', e);
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSendMessage();
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
	<div class="border-t border-base-300 p-4 pb-8 md:pb-4">
		<div class="mx-auto max-w-4xl">
			<div
				class="relative flex items-end gap-2 rounded-2xl bg-base-200 p-2 focus-within:ring-2 focus-within:ring-primary/20"
			>
				<!-- Optional: Attachment button -->
				<!-- <Button variant="ghost" circle size="sm" class="shrink-0 mb-1">
					<Paperclip class="size-5 opacity-50" />
				</Button> -->

				<Textarea
					bind:value={input}
					placeholder="Ask anything..."
					class="max-h-48 min-h-[44px] flex-1 resize-none border-none bg-transparent py-2.5 focus:ring-0"
					onkeydown={handleKeydown}
				/>

				<Button
					color="primary"
					circle
					size="sm"
					class="mb-1 shrink-0"
					disabled={!input.trim() || loading}
					onclick={handleSendMessage}
				>
					{#if loading}
						<span class="loading loading-xs loading-spinner"></span>
					{:else}
						<Send class="size-4" />
					{/if}
				</Button>
			</div>
			<div class="mt-2 text-center text-[10px] opacity-30">
				Press Enter to send, Shift+Enter for new line
			</div>
		</div>
	</div>
</div>
