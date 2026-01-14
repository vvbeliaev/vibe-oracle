import { ChatsStatusOptions, Collections, pb, type ChatsResponse } from '$lib';
import type { ListResult } from 'pocketbase';

const PAGE_SIZE = 20;

class ChatsStore {
	page = $state(1);
	totalPages = $state(0);
	totalItems = $state(0);
	loading = $state(true);

	private _chats: ChatsResponse[] = $state([]);
	private userId: string | null = null;

	chats = $derived(this._chats);

	set(chatsRes: ListResult<ChatsResponse>) {
		this.loading = false;
		this._chats = chatsRes.items;
		this.page = chatsRes.page;
		this.totalPages = chatsRes.totalPages;
		this.totalItems = chatsRes.totalItems;
	}

	async load(userId: string) {
		const res = await pb.collection(Collections.Chats).getList(1, PAGE_SIZE, {
			filter: `user = "${userId}"`,
			sort: '-created'
		});
		this.userId = userId;
		return res;
	}

	async loadNextPage() {
		if (this.page >= this.totalPages) return;

		this.loading = true;
		const res = await pb.collection(Collections.Chats).getList(this.page + 1, PAGE_SIZE, {
			filter: `user = "${this.userId}"`,
			sort: '-created'
		});
		this._chats = [...this._chats, ...res.items];
		this.page = res.page;
		this.totalPages = res.totalPages;
		this.totalItems = res.totalItems;
		this.loading = false;
	}

	getEmpty() {
		return this._chats.find((chat) => chat.status === ChatsStatusOptions.empty);
	}

	async subscribe(userId: string) {
		return pb.collection(Collections.Chats).subscribe(
			'*',
			(e) => {
				switch (e.action) {
					case 'create':
						this._chats = this._chats.filter((item) => !item.id.startsWith('temp-'));
						this._chats.unshift(e.record);
						break;
					case 'update':
						this._chats = this._chats.map((item) => (item.id === e.record.id ? e.record : item));
						break;
					case 'delete':
						this._chats = this._chats.filter((item) => item.id !== e.record.id);
						break;
				}
			},
			{ filter: `user = "${userId}"` }
		);
	}

	unsubscribe() {
		pb.collection(Collections.Chats).unsubscribe();
	}

	clear() {
		this._chats = [];
		this.page = 1;
		this.totalPages = 0;
		this.totalItems = 0;
		this.loading = true;
	}
}

export const chatsStore = new ChatsStore();
