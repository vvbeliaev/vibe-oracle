<script lang="ts">
	import { ArrowRight, GraduationCap } from 'lucide-svelte';
	import { Button, ChatsStatusOptions } from '$lib';
	import { userStore } from '$lib/apps/user';
	import { chatsStore, chatApi } from '$lib/apps/chat';
	import { goto } from '$app/navigation';

	const user = $derived(userStore.user);

	async function handleGetStarted() {
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
</script>

<div class="flex h-full flex-col items-center justify-center p-6">
	<!-- Background Glow -->
	<div class="pointer-events-none absolute inset-0 overflow-hidden">
		<div
			class="absolute top-1/2 left-1/2 size-[300px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary/10 blur-[100px]"
		></div>
	</div>

	<div class="card w-full max-w-md bg-base-200 shadow-xl lg:max-w-lg">
		<div class="card-body items-center text-center">
			<div
				class="mb-4 flex size-16 items-center justify-center rounded-2xl bg-primary text-primary-content shadow-lg shadow-primary/20"
			>
				<GraduationCap size={40} />
			</div>

			<h1 class="card-title text-3xl font-bold md:text-4xl">Welcome</h1>
			<p class="mt-2 text-lg opacity-70">
				Your personal <span class="font-semibold text-primary">Telegram Knowledge Base</span>
			</p>

			<p class="mt-4 text-sm opacity-50">
				Chat with your favorite channels, get summaries, and find answers across your community's
				shared intelligence.
			</p>

			<div class="mt-8 card-actions w-full">
				<Button size="lg" block class="h-14 rounded-xl text-lg" onclick={handleGetStarted}>
					Get Started
					<ArrowRight class="ml-2 size-5" />
				</Button>
			</div>
		</div>
	</div>
</div>

<style>
</style>
