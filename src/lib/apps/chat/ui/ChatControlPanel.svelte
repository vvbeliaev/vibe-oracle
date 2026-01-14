<script lang="ts">
	import { X } from 'lucide-svelte';

	import { Button, uiStore } from '$lib';

	interface Props {
		/** Whether the panel is in compact mode */
		compact?: boolean;
		/** Additional class for styling */
		class?: string;
	}

	let { compact = false, class: className = '' }: Props = $props();
</script>

<div class={['flex h-full flex-col', className]}>
	<header class="h-12 border-b border-base-300 px-1 mb-2 flex items-center justify-between">
		<div class={['', compact ? 'px-3' : '']}>
			<h2 class={['', compact ? 'text-base' : '']}>Control Panel</h2>
			<p class="text-sm text-base-content/70">Chat settings and widgets</p>
		</div>

		<Button class="sm:hidden" circle variant="ghost" onclick={() => uiStore.setRightSidebarOpen(false)}>
			<X size={20} />
		</Button>
	</header>

	<!-- Placeholder widgets - can be extended -->
	<div class={['space-y-3', compact ? 'px-2' : 'px-4']}>
		<div class="rounded-lg bg-base-200 p-3">
			<h3 class="text-sm font-semibold">Quick Actions</h3>
			<p class="text-xs text-base-content/60">Coming soon...</p>
		</div>

		<div class="rounded-lg bg-base-200 p-3">
			<h3 class="text-sm font-semibold">Chat Info</h3>
			<p class="text-xs text-base-content/60">Coming soon...</p>
		</div>
	</div>
</div>
