<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import type { Session } from '$lib/types';

	let {
		session,
		deleting = false,
		rememberDeleteChoice = $bindable(false),
		onCancel,
		onConfirm
	}: {
		session: Session;
		deleting?: boolean;
		rememberDeleteChoice?: boolean;
		onCancel: () => void;
		onConfirm: () => void;
	} = $props();
</script>

<div class="delete-confirm" transition:fly={{ y: 8, duration: 140 }}>
	<div class="delete-copy">
		<strong>Delete “{session.title || 'Untitled'}”?</strong>
		<span>This cannot be undone.</span>
	</div>
	<label class="delete-check">
		<input type="checkbox" bind:checked={rememberDeleteChoice} />
		<span>Don’t ask again</span>
	</label>
	<div class="delete-actions">
		<button class="cancel-delete" onclick={onCancel}>Cancel</button>
		<button class="confirm-delete" onclick={onConfirm} disabled={deleting}>Delete</button>
	</div>
</div>
<button
	class="delete-scrim"
	aria-label="Cancel delete"
	onclick={onCancel}
	transition:fade={{ duration: 100 }}
></button>

<style>
	.delete-confirm {
		position: absolute;
		left: 12px;
		right: 12px;
		bottom: 56px;
		z-index: 20;
		display: grid;
		gap: 10px;
		padding: 12px;
		border: 1px solid var(--border-soft);
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.98);
		box-shadow: var(--shadow-card);
	}

	.delete-copy {
		display: grid;
		gap: 4px;
		font-size: 12px;
		color: var(--text-muted);
	}

	.delete-copy strong {
		color: var(--text-main);
	}

	.delete-check {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 11px;
		color: var(--text-muted);
	}

	.delete-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}

	.cancel-delete,
	.confirm-delete {
		border: none;
		border-radius: 8px;
		padding: 7px 10px;
		font: inherit;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
	}

	.cancel-delete {
		background: rgba(15, 23, 42, 0.05);
		color: var(--text-main);
	}

	.confirm-delete {
		background: #b42318;
		color: white;
	}

	.confirm-delete:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.delete-scrim {
		position: absolute;
		inset: 0;
		z-index: 15;
		border: none;
		background: rgba(15, 23, 42, 0.12);
		cursor: pointer;
	}
</style>
