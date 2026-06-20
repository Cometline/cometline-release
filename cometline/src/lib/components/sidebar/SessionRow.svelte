<script lang="ts">
	import { Trash2 } from '@lucide/svelte';
	import type { Session } from '$lib/types';

	let {
		session,
		selected = false,
		deleting = false,
		onSelect,
		onDelete
	}: {
		session: Session;
		selected?: boolean;
		deleting?: boolean;
		onSelect: () => void;
		onDelete: () => void;
	} = $props();
</script>

<div class="session-row-wrap" class:selected>
	<button class="session-row" onclick={onSelect}>
		<span class="session-title">{session.title || 'Untitled'}</span>
	</button>
	<button
		class="delete-session"
		disabled={deleting}
		onclick={onDelete}
		aria-label={`Delete ${session.title || 'Untitled'}`}
		title="Delete session"
	>
		<Trash2 size={13} stroke-width={1.9} />
	</button>
</div>

<style>
	.session-row-wrap {
		position: relative;
		display: flex;
		align-items: stretch;
		border-radius: 8px;
	}

	.session-row-wrap:hover {
		background: rgba(0, 0, 0, 0.08);
	}

	.session-row-wrap.selected {
		background: rgba(0, 0, 0, 0.06);
	}

	.session-row {
		width: 100%;
		text-align: left;
		padding: 7px 10px;
		padding-right: 34px;
		border-radius: 8px;
		border: none;
		background: transparent;
		color: var(--text-main);
		font-size: 13px;
		line-height: 1.35;
		font-weight: 450;
		cursor: pointer;
	}

	.session-title {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		display: block;
	}

	.delete-session {
		position: absolute;
		right: 5px;
		top: 50%;
		transform: translateY(-50%);
		width: 24px;
		height: 24px;
		border: none;
		border-radius: 6px;
		background: transparent;
		color: var(--text-soft);
		display: grid;
		place-items: center;
		opacity: 0;
		cursor: pointer;
	}

	.session-row-wrap:hover .delete-session,
	.session-row-wrap:focus-within .delete-session {
		opacity: 1;
	}

	.delete-session:hover:not(:disabled) {
		background: rgba(180, 35, 24, 0.08);
		color: #b42318;
	}

	.delete-session:disabled {
		opacity: 0.35;
	}
</style>
