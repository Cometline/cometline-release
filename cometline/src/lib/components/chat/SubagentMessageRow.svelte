<script lang="ts">
	import SubagentPanel from '$lib/components/chat/SubagentPanel.svelte';
	import ThreadAvatar from '$lib/components/chat/ThreadAvatar.svelte';
	import { startsSpeakerRun } from '$lib/conversation/thread-view-helpers';
	import type { AssistantStackFoldController } from '$lib/conversation/assistant-stack-props';
	import type { ChatItem } from '$lib/stores/chat.svelte';
	import type { IconVariant } from '$lib/types';

	let {
		item,
		threadItems,
		index,
		iconVariant,
		fold
	}: {
		item: Extract<ChatItem, { type: 'subagent' }>;
		threadItems: readonly ChatItem[];
		index: number;
		iconVariant: IconVariant;
		fold: AssistantStackFoldController;
	} = $props();
</script>

<div
	class="row tool-row subagent-row"
	class:continuation-row={!startsSpeakerRun(threadItems, index, 'assistant')}
>
	<ThreadAvatar variant="gutter" {iconVariant} />
	<div class="subagent-stack">
		<SubagentPanel
			{item}
			expanded={fold.subagentExpanded(item.id)}
			onToggle={() => fold.toggleSubagent(item.id)}
		/>
	</div>
</div>

<style>
	.row {
		display: flex;
		width: 100%;
		gap: var(--chat-row-gap);
	}

	.continuation-row {
		margin-top: -16px;
	}

	.tool-row {
		justify-content: flex-start;
	}

	.subagent-row {
		align-items: flex-start;
	}

	.subagent-stack {
		min-width: 0;
		flex: 1;
		max-width: var(--chat-assistant-column);
	}
</style>
