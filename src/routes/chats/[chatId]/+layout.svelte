<script lang="ts">
	import { page } from '$app/state';
	import { messagesStore } from '$lib/apps/chat';

	const { children } = $props();

	const chatId = $derived(page.params.chatId);

	$effect(() => {
		if (!chatId) return;

		messagesStore.load(chatId).then((res) => {
			messagesStore.set(res);
			messagesStore.subscribe(chatId);
		});
		return () => {
			messagesStore.unsubscribe();
		};
	});
</script>

<div class="h-full w-full overflow-hidden">
	{@render children()}
</div>
