export type MessageChunk = {
	text: string;
	msgId: string;
	i?: number;
};

export type Sender = {
	id: string;
	avatar: string;
	name: string;
	role: string;
};

export type Citation = {
	id: string;
	link: string;
	snippet: string;
};

export type ChatResponse = {
	messageId: string;
	content: string;
	citations: Citation[];
};
