<script lang="ts">
	import posthog from 'posthog-js';
	import {
		House,
		Search,
		CalendarDays,
		Settings,
		LogIn,
		PanelRight,
		Menu,
		Plus,
		MessageSquare
	} from 'lucide-svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';

	import '$lib/shared/pb/pb-hook';
	import {
		ThemeLoad,
		PortalHost,
		Logo,
		Button,
		ThemeController,
		swipeable,
		uiStore,
		Sidebar,
		Collections,
		ChatsStatusOptions
	} from '$lib';
	import { userStore } from '$lib/apps/user';
	import { chatsStore, chatApi } from '$lib/apps/chat';

	import './layout.css';
	import PWA from './PWA.svelte';
	import Splash from './Splash.svelte';

	const nav = [{ label: 'Home', href: '/', icon: House }];

	let { children, data } = $props();
	const globalPromise = $derived(data.globalPromise);

	const user = $derived(userStore.user);
	const chats = $derived(chatsStore.chats);

	const isActive = (path: string) => page.url.pathname === path;

	let loaderRef: HTMLElement | undefined = $state();

	$effect(() => {
		if (!loaderRef || chatsStore.page >= chatsStore.totalPages) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && !chatsStore.loading) {
					chatsStore.loadNextPage();
				}
			},
			{ threshold: 0.1 }
		);

		observer.observe(loaderRef);
		return () => observer.disconnect();
	});

	async function handleNewChat() {
		if (!user) {
			goto('/auth');
			return;
		}

		let chat = chatsStore.getEmpty();
		if (!chat) {
			chat = await chatApi.create({
				user: user.id,
				title: 'New Chat',
				status: ChatsStatusOptions.empty
			});
		}
		goto(`/chats/${chat.id}`);
	}

	// Posthog identify and set person
	$effect(() => {
		console.log(user);

		if (!user) return;

		posthog.identify(user.id, {
			email: user.email,
			name: user.name
		});
		posthog.capture('user_authenticated', {
			email: user.email,
			name: user.name
		});
	});

	// Global user load
	$effect(() => {
		globalPromise.then(({ userAuth, chatsRes }) => {
			if (userAuth) userStore.set(userAuth);
			if (chatsRes) chatsStore.set(chatsRes);
		});
	});

	// Real-time subscription
	$effect(() => {
		const userId = userStore.user?.id;
		if (!userId) return;
		userStore.subscribe();
		chatsStore.subscribe(userId);
		return () => {
			userStore.unsubscribe();
			chatsStore.unsubscribe();
		};
	});

	afterNavigate(() => {
		uiStore.setSidebarOpen(false);
	});
</script>

<PWA />

<svelte:head>
	<link rel="icon" type="image/x-icon" href="/favicon_io/favicon.ico" />
	<link rel="apple-touch-icon" sizes="180x180" href="/favicon_io/apple-touch-icon.png" />
	<link rel="icon" type="image/png" sizes="32x32" href="/favicon_io/favicon-32x32.png" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicon_io/favicon-16x16.png" />
	<ThemeLoad />
</svelte:head>

{#snippet sidebarHeader({ expanded }: { expanded: boolean })}
	{#if expanded}
		<a href="/" class="flex items-center gap-2">
			<Logo />
		</a>
	{/if}
{/snippet}

{#snippet sidebarContent({ expanded }: { expanded: boolean })}
	<div class="shrink-0 px-2 pt-4">
		<Button
			block
			class="rounded-xl"
			disabled={chatsStore.loading}
			square={!expanded}
			onclick={handleNewChat}
		>
			<Plus class="size-5" />
			{#if expanded}
				<span class="text-nowrap">New Chat</span>
			{/if}
		</Button>
	</div>

	<div class="divider my-2"></div>

	<div class="flex-1 overflow-y-auto px-2">
		<ul class="menu w-full gap-1">
			{#each chats as chat (chat.id)}
				<li class="w-full">
					<a
						href="/chats/{chat.id}"
						class={[
							'btn flex w-full items-center gap-2 rounded-xl btn-ghost transition-all',
							expanded ? 'justify-start px-4' : 'justify-center',
							isActive(`/chats/${chat.id}`) ? 'btn-soft' : ''
						]}
						title={!expanded ? chat.title || chat.id : ''}
					>
						{chat.id.slice(0, 2)}.
						{#if expanded}
							<span class="truncate font-medium">{chat.title || chat.id}</span>
						{/if}
					</a>
				</li>
			{/each}
		</ul>

		<!-- Infinite scroll loader -->
		{#if chatsStore.page < chatsStore.totalPages}
			<div bind:this={loaderRef} class="flex justify-center py-4">
				<span class="loading loading-sm loading-spinner"></span>
			</div>
		{/if}
	</div>
{/snippet}

{#snippet sidebarFooter({ expanded }: { expanded: boolean })}
	<!-- <div class="divider my-1"></div> -->

	{#if user && user.verified}
		<div class="mb-1 flex justify-center px-2">
			<!-- <button
				class={['btn justify-start btn-ghost', expanded ? 'btn-block' : 'btn-square']}
				onclick={() => uiStore.toggleFeedbackModal()}
			>
				<MessageSquare class={expanded ? 'size-5' : 'size-6'} />
				{#if expanded}
					Feedback
				{:else}
					<span class="sr-only">Feedback</span>
				{/if}
			</button> -->
		</div>
	{/if}

	<div class={['mb-1 border-base-300', expanded ? 'px-2' : 'flex justify-center']}>
		<ThemeController {expanded} navStyle />
	</div>

	<div class="border-t border-base-300">
		{#if user && user.verified}
			<a
				href="/profile"
				class={[
					'flex items-center gap-3 p-2 transition-colors hover:bg-base-200',
					!expanded && 'justify-center'
				]}
				title={!expanded ? 'Settings' : ''}
			>
				{#if userStore.avatarUrl}
					<img src={userStore.avatarUrl} alt={user.name} class="size-10 rounded-full" />
				{:else}
					<div class="flex size-10 items-center justify-center rounded-full bg-base-300">
						{user.name?.at(0)?.toUpperCase() ?? 'U'}
					</div>
				{/if}
				{#if expanded}
					<div class="flex-1 overflow-hidden">
						<div class="truncate text-sm font-semibold">{user.name || '<No Name>'}</div>
						<div class="truncate text-xs opacity-60">{user.email}</div>
					</div>
					<Settings class="size-5 opacity-60" />
				{/if}
			</a>
		{:else}
			<a
				href="/auth"
				class={[
					'flex items-center gap-3 rounded-lg p-2 transition-colors hover:bg-base-300',
					!expanded && 'justify-center'
				]}
				title={!expanded ? 'Log in' : ''}
			>
				<div class="size-10 rounded-full bg-base-300"></div>
				{#if expanded}
					<div class="flex-1 overflow-hidden">
						<div class="truncate text-sm font-semibold">Log in</div>
					</div>
				{/if}
			</a>
		{/if}
	</div>
{/snippet}

{#await globalPromise}
	<Splash />
{:then}
	<div
		class="flex h-screen flex-col overflow-hidden bg-base-100 md:flex-row"
		use:swipeable={{
			isOpen: uiStore.sidebarOpen ?? false,
			direction: 'right',
			onOpen: () => uiStore.setSidebarOpen(true),
			onClose: () => uiStore.setSidebarOpen(false)
		}}
	>
		<!-- Sidebar -->
		<Sidebar
			open={uiStore.sidebarOpen ?? false}
			expanded={uiStore.sidebarExpanded ?? true}
			position="left"
			header={sidebarHeader}
			footer={sidebarFooter}
			onclose={() => uiStore.setSidebarOpen(false)}
			ontoggle={() => uiStore.toggleSidebarExpanded()}
		>
			{#snippet children({ expanded })}
				{@render sidebarContent({ expanded })}
			{/snippet}
		</Sidebar>

		<!-- Main Content -->
		<main class="mb-12 flex-1 overflow-hidden md:mb-0">
			<div class="mx-auto h-full max-w-[1440px]">
				{@render children()}
			</div>
		</main>

		<!-- Mobile Dock -->
		<div class="dock dock-sm border-t border-base-300 md:hidden">
			{#each nav as item}
				<a href={item.href} class:dock-active={page.url.pathname === item.href}>
					<item.icon class="size-5" />
					<span class="dock-label">{item.label}</span>
				</a>
			{/each}

			{#if user && user.verified}
				<a href="/profile" class:dock-active={page.url.pathname === '/profile'}>
					<Settings class="size-5" />
					<span class="dock-label">Profile</span>
				</a>
			{:else}
				<a href="/auth" class:dock-active={page.url.pathname === '/auth'}>
					<LogIn class="size-5" />
					<span class="dock-label">Log In</span>
				</a>
			{/if}

			<!-- Hidden for now -->
			<button class="hidden" onclick={() => uiStore.toggleRightSidebar()}>
				<PanelRight class="size-5" />
				<span class="dock-label">Panel</span>
			</button>
			<button class="hidden" onclick={() => uiStore.setSidebarOpen(true)}>
				<Menu class="size-5" />
				<span class="dock-label">Menu</span>
			</button>
		</div>
	</div>
{/await}

<PortalHost />
