/**
 * Serializes chat turns so only one startChat/send runs at a time.
 * Additional submits while busy are queued FIFO and drained automatically.
 */

import type { ImageAttachment } from '$lib/types';

export interface QueuedMessage {
	id: string;
	text: string;
	images?: ImageAttachment[];
}

export interface ChatTurnQueue {
	enqueue(text: string, images?: ImageAttachment[]): Promise<boolean>;
	remove(id: string): boolean;
	clear(): void;
	readonly pendingCount: number;
	readonly pendingMessages: readonly QueuedMessage[];
	readonly processing: boolean;
}

export function createChatTurnQueue(
	runTurn: (text: string, images?: ImageAttachment[]) => Promise<void>,
	onChange?: () => void
): ChatTurnQueue {
	let queue: QueuedMessage[] = [];
	let processing = false;
	let activeTurnText: string | null = null;
	let nextID = 0;

	function notifyChange() {
		onChange?.();
	}

	function createQueuedMessage(text: string, images?: ImageAttachment[]): QueuedMessage {
		nextID += 1;
		return { id: `queued-${Date.now()}-${nextID}`, text, images };
	}

	function signature(text: string, images?: ImageAttachment[]): string {
		return JSON.stringify({ text, images: images?.map((image) => image.data) ?? [] });
	}

	function isDuplicate(text: string, images?: ImageAttachment[]): boolean {
		const sig = signature(text, images);
		if (activeTurnText === sig) return true;
		const last = queue.at(-1);
		return last ? signature(last.text, last.images) === sig : false;
	}

	function queueTurn(text: string, images?: ImageAttachment[]): boolean {
		if (isDuplicate(text, images)) return false;
		queue.push(createQueuedMessage(text, images));
		notifyChange();
		return true;
	}

	async function runTurnTracked(text: string, images?: ImageAttachment[]): Promise<void> {
		activeTurnText = signature(text, images);
		try {
			if (images === undefined) {
				await runTurn(text);
			} else {
				await runTurn(text, images);
			}
		} finally {
			if (activeTurnText === signature(text, images)) activeTurnText = null;
		}
	}

	async function runLoop(initialText?: string, initialImages?: ImageAttachment[]): Promise<boolean> {
		if (processing) {
			if (initialText !== undefined) return queueTurn(initialText, initialImages);
			return false;
		}

		processing = true;
		notifyChange();
		try {
			if (initialText !== undefined) {
				await runTurnTracked(initialText, initialImages);
			}
			while (queue.length > 0) {
				const { text, images } = queue.shift()!;
				notifyChange();
				await runTurnTracked(text, images);
			}
		} finally {
			activeTurnText = null;
			processing = false;
			notifyChange();
		}
		return true;
	}

	return {
		get pendingCount() {
			return queue.length;
		},
		get pendingMessages() {
			return queue;
		},
		get processing() {
			return processing;
		},
		enqueue(text: string, images?: ImageAttachment[]) {
			if (processing) return Promise.resolve(queueTurn(text, images));
			return runLoop(text, images);
		},
		remove(id: string) {
			const index = queue.findIndex((item) => item.id === id);
			if (index < 0) return false;
			queue.splice(index, 1);
			notifyChange();
			return true;
		},
		clear() {
			queue = [];
			notifyChange();
		}
	};
}
