<script lang="ts">
	// @ts-ignore
	import { DateTime } from 'luxon';
	import { marked } from 'marked';
	import DOMPurify from 'dompurify';
	import type { ClassValue } from 'svelte/elements';
	import { ExternalLink } from 'lucide-svelte';

	import type { Sender, Citation } from '$lib/apps/chat';
	import type { MessagesResponse } from '$lib';

	interface Props {
		class?: ClassValue;
		incoming: boolean;
		msg: MessagesResponse<{ citations?: Citation[] }>;
		sender: Sender;
		showHeader?: boolean;
	}

	const { msg, incoming, class: className = '', sender, showHeader = true }: Props = $props();

	// TIME
	const utcTs = $derived(
		DateTime.fromFormat(msg.created || '', "yyyy-MM-dd HH:mm:ss.SSS'Z'", {
			zone: 'utc'
		})
	);

	const localTs = $derived(utcTs.isValid ? utcTs.toLocal() : utcTs);
	const formattedTime = $derived(localTs.isValid ? localTs.toFormat('h:mm a') : '');

	const rawHtml = $derived(marked.parse(msg.content || ''));
	const safeHtml = $derived(
		DOMPurify.sanitize(rawHtml as string, {
			ADD_ATTR: ['target', 'rel']
		})
	);

	const isWaitingForResponse = $derived(
		incoming && msg.content.trim() === '' && msg.status === 'streaming'
	);

	const citations = $derived(msg.meta?.citations || []);
</script>

<div class={['group flex w-full gap-4 px-4 py-2', incoming ? '' : 'flex-row-reverse', className]}>
	<!-- Avatar -->
	<div class="flex shrink-0 flex-col items-center gap-1">
		{#if showHeader}
			{#if sender.avatar}
				<div class="avatar">
					<div class="size-10 overflow-hidden rounded-full ring-1 ring-base-300 ring-offset-1">
						<img alt={msg.role} src={sender.avatar} class="h-full w-full object-cover" />
					</div>
				</div>
			{:else}
				<div class="flex size-10 items-center justify-center rounded-full bg-base-300 text-center">
					{sender.name?.at(0)?.toUpperCase() ?? 'U'}
				</div>
			{/if}
			<!-- {:else}
			<div class="size-10"></div> -->
		{/if}
	</div>

	<!-- Message Content -->
	<div class={['flex max-w-full flex-col', incoming ? 'items-start' : 'items-end']}>
		<!-- Header (Name) -->
		{#if showHeader}
			<div class="mb-1 flex items-center gap-2 px-1">
				<span class="text-sm font-semibold text-base-content">
					{sender?.name || 'Unknown'}
				</span>
				{#if formattedTime}
					<time
						datetime={msg.created}
						class="text-xs text-base-content/40 opacity-80 transition-opacity group-hover:opacity-100"
					>
						{formattedTime}
					</time>
				{/if}
			</div>
		{/if}

		<!-- Bubble -->
		<div
			class={[
				'relative prose overflow-hidden rounded-2xl px-3 py-3 text-base leading-relaxed shadow-sm transition-all',
				incoming
					? 'rounded-tl-none bg-base-200 text-base-content'
					: 'rounded-tr-none bg-primary text-primary-content',

				// текущие стили:
				'[&_p]:m-0 [&_p]:min-h-[1em]',
				'[&_a]:underline [&_a]:decoration-current/30 [&_a]:underline-offset-2 hover:[&_a]:decoration-current',
				'[&_code]:rounded-md [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:font-mono [&_code]:text-sm',

				// ВАЖНО: отдельная настройка для блочных код-блоков
				// <pre> + <code> внутри
				'[&_pre]:my-2 [&_pre]:max-w-full [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:p-3',
				'[&_pre]:bg-base-300 [&_pre]:text-base-content', // нормальный контраст
				'[&_pre_code]:bg-transparent [&_pre_code]:text-base-content'
			]}
		>
			{#if isWaitingForResponse}
				<div class="flex items-center gap-1 py-2">
					<span
						class="h-2 w-2 animate-bounce rounded-full bg-current opacity-60"
						style="animation-delay: 0ms"
					></span>
					<span
						class="h-2 w-2 animate-bounce rounded-full bg-current opacity-60"
						style="animation-delay: 150ms"
					></span>
					<span
						class="h-2 w-2 animate-bounce rounded-full bg-current opacity-60"
						style="animation-delay: 300ms"
					></span>
				</div>
			{:else}
				{@html safeHtml}
			{/if}

			{#if citations.length > 0}
				<div class="mt-4 border-t border-base-content/10 pt-3">
					<div class="mb-2 text-[10px] font-bold tracking-wider uppercase opacity-40">Sources</div>
					<div class="flex flex-wrap gap-2">
						{#each citations as citation, i}
							<a
								href={citation.link}
								target="_blank"
								rel="noopener noreferrer"
								class="flex items-center gap-1.5 rounded-full bg-base-300/50 px-2.5 py-1 text-xs transition-colors hover:bg-base-300"
								title={citation.snippet}
							>
								<span class="font-medium">[{i + 1}]</span>
								<span class="max-w-[120px] truncate opacity-70">
									{citation.link.split('/').pop() || 'Source'}
								</span>
								<ExternalLink class="size-3 opacity-40" />
							</a>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>
