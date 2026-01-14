import { Collections, nanoid, pb } from '$lib';

import { chatsStore } from '$lib/apps/chat/chats.svelte';

import { userStore } from './user.svelte';

export async function globalUserLoad() {
	console.log('globalUserLoad', pb.authStore.isValid);

	if (!pb.authStore.isValid) {
		try {
			const userAuth = await authGuest();
			const chatsRes = await chatsStore.load(userAuth.record.id);
			return { userAuth, chatsRes };
		} catch (error) {
			console.error(error);
			pb.authStore.clear();
			return { userAuth: null, chatsRes: null };
		}
	}

	try {
		const userAuth = await userStore.load();
		const chatsRes = await chatsStore.load(userAuth.record.id);
		return { userAuth, chatsRes };
	} catch (error) {
		console.error(error);
		pb.authStore.clear();
		return { userAuth: null, chatsRes: null };
	}
}

async function authGuest() {
	let guestId = localStorage.getItem('guest_id') ?? '';
	let randomPassword = localStorage.getItem('guest_password') ?? '';

	if (!guestId || !randomPassword) {
		guestId = nanoid();
		randomPassword = nanoid();
		await pb.collection(Collections.Users).create({
			guest: guestId,
			password: randomPassword,
			passwordConfirm: randomPassword
		});
	}
	localStorage.setItem('guest_id', guestId);
	localStorage.setItem('guest_password', randomPassword);

	const authRes = await pb.collection(Collections.Users).authWithPassword(guestId, randomPassword);
	return authRes;
}
