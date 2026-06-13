/**
 * Serializes chat turns so only one startChat/send runs at a time.
 * Additional submits while busy are queued FIFO and drained automatically.
 */

export interface QueuedMessage {
	id: string;
	text: string;
}

export interface ChatTurnQueue {
	enqueue(text: string): Promise<void>;
	remove(id: string): boolean;
	clear(): void;
	readonly pendingCount: number;
	readonly pendingMessages: readonly QueuedMessage[];
	readonly processing: boolean;
}

export function createChatTurnQueue(
	runTurn: (text: string) => Promise<void>,
	onChange?: () => void
): ChatTurnQueue {
	let queue: QueuedMessage[] = [];
	let processing = false;
	let nextID = 0;

	function notifyChange() {
		onChange?.();
	}

	function createQueuedMessage(text: string): QueuedMessage {
		nextID += 1;
		return { id: `queued-${Date.now()}-${nextID}`, text };
	}

	async function runLoop(initialText?: string) {
		if (processing && initialText === undefined) return;
		processing = true;
		notifyChange();
		try {
			if (initialText !== undefined) {
				await runTurn(initialText);
			}
			while (queue.length > 0) {
				const { text } = queue.shift()!;
				notifyChange();
				await runTurn(text);
			}
		} finally {
			processing = false;
			notifyChange();
		}
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
		enqueue(text: string) {
			if (processing) {
				queue.push(createQueuedMessage(text));
				notifyChange();
				return Promise.resolve();
			}
			return runLoop(text);
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
