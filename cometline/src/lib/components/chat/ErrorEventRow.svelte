<script lang="ts">
	import { TriangleAlert } from '@lucide/svelte';
	import type { ChatItem } from '$lib/stores/chat.svelte';

	let { item }: { item: Extract<ChatItem, { type: 'error' }> } = $props();
</script>

<div class="row event-row">
	<div class="event-card error-card">
		<div class="event-title">
			<TriangleAlert size={14} /><span>Error</span>
		</div>
		<p>{item.text}</p>
	</div>
</div>

<style>
	.row {
		display: flex;
		width: 100%;
		gap: var(--chat-row-gap);
	}

	.event-row {
		justify-content: flex-start;
	}

	.event-row .event-card {
		max-width: var(--chat-content-column);
	}

	.event-card {
		min-width: 0;
		width: 100%;
		max-width: 100%;
		border: 1px solid var(--border-soft);
		background: rgba(255, 255, 255, 0.68);
		border-radius: 14px;
		padding: 10px 12px;
		color: var(--text-muted);
		box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
	}

	.event-title {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 12px;
		font-weight: 650;
		color: var(--text-main);
		margin-bottom: 7px;
	}

	.event-title :global(svg:last-child) {
		flex-shrink: 0;
	}

	.error-card p {
		margin: 0;
		font-size: 12px;
		line-height: 1.5;
		white-space: pre-wrap;
	}

	.error-card {
		color: #b42318;
		background: rgba(255, 245, 245, 0.82);
		border-color: rgba(244, 63, 94, 0.18);
	}
</style>
