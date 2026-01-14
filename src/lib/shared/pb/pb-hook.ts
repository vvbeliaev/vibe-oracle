import type { AuthRecord } from 'pocketbase';

import { pb, setPBCookie, type UsersResponse } from '$lib';
import { userStore } from '$lib/apps/user';
import { chatsStore, messagesStore } from '$lib/apps/chat';

pb.authStore.onChange((token: string, record: AuthRecord) => {
	if (record && pb!.authStore.isValid) {
		try {
			const user = record as UsersResponse;
			userStore.set({ record: user, token });

			setPBCookie();
		} catch (error) {
			userStore.clear();
			chatsStore.clear();
			messagesStore.clear();
			console.error('Failed to parse user data:', error);
		}
	} else {
		userStore.clear();
	}
}, false);
