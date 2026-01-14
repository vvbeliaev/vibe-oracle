<script lang="ts">
	import { MediaQuery } from 'svelte/reactivity';
	import { page } from '$app/state';

	import { messagesStore, ChatControlPanel } from '$lib/apps/chat';
	import { uiStore, Sidebar, swipeable, Button } from '$lib/shared/ui';

	const { children } = $props();

	const chatId = $derived(page.params.chatId);
	const rightSidebarOpen = $derived(uiStore.rightSidebarOpen);

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

	const mobile = $derived(new MediaQuery('(max-width: 768px)'));
</script>

<div
	class="flex h-full w-full"
	use:swipeable={{
		isOpen: rightSidebarOpen ?? false,
		direction: 'left',
		edgeWidth: 30,
		onOpen: () => uiStore.setRightSidebarOpen(true),
		onClose: () => uiStore.setRightSidebarOpen(false)
	}}
>
	<div class="h-full flex-4 overflow-hidden">
		{@render children()}
	</div>

	<!-- Desktop Right Sidebar (always visible on desktop) -->
	<aside class="hidden w-84 flex-5 shrink-0 border-x border-base-300 md:flex md:flex-col">
		<ChatControlPanel />
	</aside>

	<!-- Mobile Right Sidebar Drawer -->
	<Sidebar
		open={(mobile.current && rightSidebarOpen) ?? false}
		position="right"
		mobileWidth="w-96 max-w-[calc(100vw-3rem)]"
		showToggle={false}
		mobileOnly
		onclose={() => uiStore.setRightSidebarOpen(false)}
	>
		{#snippet children({ expanded })}
			<ChatControlPanel compact />
		{/snippet}
	</Sidebar>
</div>
